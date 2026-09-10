(function () {
  'use strict';

  var words = [];
  var blockIds = [];
  var sentenceIds = [];

  var activeIndex = -1;
  var anchorIndex = -1;
  var marking = false;
  var activeEl = null;

  var linesCache = null;
  var cachedScrollHeight = -1;
  var cachedWidth = -1;

  var badge = null;
  var badgeTimer = null;
  var colorStyle = null;

  var searchMatches = [];
  var searchIndex = -1;

  // Containers whose text is never wrapped into navigable words: code
  // (fenced + inline), rendered math/diagrams, the line-number gutter, the
  // frontmatter block, task-list checkboxes, image fallbacks, and the nav
  // chrome itself.
  var EXCLUDED = 'pre, code, .katex, .katex-display, .mermaid, ' +
    '.hljs-ln-numbers, .hljs-ln-n, .o-mark-frontmatter, ' +
    '.task-list-item-checkbox, .img-external, img, svg, script, style, ' +
    '.ow-nav-badge';

  var BLOCK_TAGS = {
    P: 1, H1: 1, H2: 1, H3: 1, H4: 1, H5: 1, H6: 1, LI: 1, DT: 1, DD: 1,
    BLOCKQUOTE: 1, TD: 1, TH: 1, PRE: 1, FIGCAPTION: 1
  };

  var ABBREVS = {
    'e.g.': 1, 'i.e.': 1, 'etc.': 1, 'vs.': 1, 'mr.': 1, 'mrs.': 1, 'ms.': 1,
    'dr.': 1, 'st.': 1, 'prof.': 1, 'jr.': 1, 'sr.': 1, 'approx.': 1,
    'fig.': 1, 'cf.': 1, 'inc.': 1, 'ltd.': 1, 'no.': 1, 'vol.': 1, 'ed.': 1
  };

  function getBlock(el) {
    var n = el.parentElement;
    while (n) {
      if (BLOCK_TAGS[n.tagName]) return n;
      if (n.classList) {
        if (n.classList.contains('admonition-title') ||
            n.classList.contains('o-mark-frontmatter-content')) return n;
      }
      n = n.parentElement;
    }
    return document.body;
  }

  function endsSentence(text) {
    if (!/[.!?]$/.test(text)) return false;
    if (ABBREVS[text.toLowerCase()]) return false;
    if (/^\d+(\.\d+)?[.!?]$/.test(text)) return false; // number, e.g. "3.14"
    if (/^[A-Z]\.$/.test(text)) return false; // single-letter initial
    return text.length > 2;
  }

  function wrapTextNode(node) {
    var text = node.nodeValue;
    var frag = document.createDocumentFragment();
    var re = /\S+/g;
    var m;
    var last = 0;
    while ((m = re.exec(text)) !== null) {
      if (m.index > last) {
        frag.appendChild(document.createTextNode(text.slice(last, m.index)));
      }
      var span = document.createElement('span');
      span.className = 'ow-word';
      span.textContent = m[0];
      words.push(span);
      frag.appendChild(span);
      last = m.index + m[0].length;
    }
    if (last < text.length) {
      frag.appendChild(document.createTextNode(text.slice(last)));
    }
    if (node.parentNode) node.parentNode.replaceChild(frag, node);
  }

  function wrapWords() {
    words = [];
    var walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT, null);
    var textNodes = [];
    var n;
    while ((n = walker.nextNode())) {
      if (!/\S/.test(n.nodeValue)) continue;
      var parent = n.parentElement;
      if (!parent) continue;
      if (parent.matches(EXCLUDED) || parent.closest(EXCLUDED)) continue;
      textNodes.push(n);
    }
    for (var i = 0; i < textNodes.length; i++) wrapTextNode(textNodes[i]);
  }

  function assignStructure() {
    blockIds = new Array(words.length);
    sentenceIds = new Array(words.length);
    var lastBlock = null;
    var bId = -1;
    for (var i = 0; i < words.length; i++) {
      var b = getBlock(words[i]);
      if (b !== lastBlock) { bId++; lastBlock = b; }
      blockIds[i] = bId;
    }
    var sId = 0;
    var needNew = true;
    for (var j = 0; j < words.length; j++) {
      if (needNew) { sId++; needNew = false; }
      sentenceIds[j] = sId;
      if (endsSentence(words[j].textContent)) needNew = true;
    }
  }

  // Replicates Qt.lighter()/Qt.darker(): both scale the HSV V channel by a
  // factor (Qt.darker(c, f) === Qt.lighter(c, 1/f)), not an additive RGB
  // shift like CSS lighten()/darken(). Needed so the badge matches the
  // "Reloaded" badge in main.qml, which derives its colors this way.
  function rgbToHsv(r, g, b) {
    r /= 255; g /= 255; b /= 255;
    var max = Math.max(r, g, b), min = Math.min(r, g, b);
    var h, s, v = max;
    var d = max - min;
    s = max === 0 ? 0 : d / max;
    if (max === min) {
      h = 0;
    } else {
      switch (max) {
        case r: h = (g - b) / d + (g < b ? 6 : 0); break;
        case g: h = (b - r) / d + 2; break;
        default: h = (r - g) / d + 4;
      }
      h /= 6;
    }
    return { h: h, s: s, v: v };
  }

  function hsvToRgb(h, s, v) {
    var r, g, b;
    var i = Math.floor(h * 6);
    var f = h * 6 - i;
    var p = v * (1 - s);
    var q = v * (1 - f * s);
    var t = v * (1 - (1 - f) * s);
    switch (i % 6) {
      case 0: r = v; g = t; b = p; break;
      case 1: r = q; g = v; b = p; break;
      case 2: r = p; g = v; b = t; break;
      case 3: r = p; g = q; b = v; break;
      case 4: r = t; g = p; b = v; break;
      default: r = v; g = p; b = q;
    }
    return { r: Math.round(r * 255), g: Math.round(g * 255), b: Math.round(b * 255) };
  }

  function qtLighter(rgb, factor) {
    var hsv = rgbToHsv(rgb.r, rgb.g, rgb.b);
    var v = hsv.v === 0 ? Math.min(1, 1 - 1 / factor) : Math.min(1, hsv.v * factor);
    return hsvToRgb(hsv.h, hsv.s, v);
  }

  function qtDarker(rgb, factor) {
    return qtLighter(rgb, 1 / factor);
  }

  function parseRgb(str) {
    var m = /rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/.exec(str || '');
    if (!m) return { r: 255, g: 255, b: 255 };
    return { r: +m[1], g: +m[2], b: +m[3] };
  }

  function applyBadgeColors() {
    if (!badge) return;
    var bg = parseRgb(getComputedStyle(document.body).backgroundColor);
    var lum = 0.299 * bg.r / 255 + 0.587 * bg.g / 255 + 0.114 * bg.b / 255;
    var badgeBg, fg;
    if (lum < 0.5) {
      badgeBg = qtLighter(bg, 3.0);
      fg = '#d0d0d0';
    } else {
      badgeBg = qtDarker(bg, 1.5);
      fg = '#303030';
    }
    badge.style.setProperty('--ow-badge-bg', 'rgb(' + badgeBg.r + ',' + badgeBg.g + ',' + badgeBg.b + ')');
    badge.style.setProperty('--ow-badge-fg', fg);
  }

  function themeVar(name, fallback) {
    var v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
    return v !== '' ? v : fallback;
  }

  // Keyboard marking (::highlight) and mouse dragging (::selection) share one
  // rule so both gestures look the same. The color is set explicitly because
  // this rule only overrides the declarations it repeats: leaving it out would
  // keep a theme's own ::selection color (writer.css turns the text white).
  // nav.css is injected after the theme CSS, so this wins on equal specificity.
  function applyHighlightColors() {
    if (!colorStyle) {
      colorStyle = document.createElement('style');
      document.head.appendChild(colorStyle);
    }
    var accent = themeVar('--o-mark-accent', '#58a6ff');
    var markBg = themeVar('--o-mark-nav-mark-bg', 'color-mix(in srgb, ' + accent + ' 25%, transparent)');
    var markFg = themeVar('--o-mark-nav-mark-fg', getComputedStyle(document.body).color);
    var searchBg = themeVar('--o-mark-nav-search-bg', 'color-mix(in srgb, ' + accent + ' 35%, transparent)');
    var searchCurrentBg = themeVar('--o-mark-nav-search-current-bg', accent);
    var searchCurrentFg = themeVar('--o-mark-nav-search-current-fg', getComputedStyle(document.body).backgroundColor);
    colorStyle.textContent =
      '::selection, ::highlight(ow-mark) {' +
      'background-color: ' + markBg + ';' +
      'color: ' + markFg + ';' +
      '}' +
      '::highlight(ow-search) {' +
      'background-color: ' + searchBg + ';' +
      '}' +
      '::highlight(ow-search-current) {' +
      'background-color: ' + searchCurrentBg + ';' +
      'color: ' + searchCurrentFg + ';' +
      '}';
  }

  function setActive(idx) {
    if (idx < 0) idx = 0;
    if (idx >= words.length) idx = words.length - 1;
    if (activeEl) activeEl.classList.remove('ow-active');
    activeIndex = idx;
    activeEl = words[idx];
    activeEl.classList.add('ow-active');
    if (marking) updateRange();
  }

  function scrollActiveIntoView() {
    if (!activeEl) return;
    var r = activeEl.getBoundingClientRect();
    var margin = 40;
    if (r.top < margin) window.scrollBy(0, r.top - margin);
    else if (r.bottom > window.innerHeight - margin) {
      window.scrollBy(0, r.bottom - (window.innerHeight - margin));
    }
  }

  function ensureLines() {
    var sh = document.documentElement.scrollHeight;
    var w = window.innerWidth;
    if (linesCache && sh === cachedScrollHeight && w === cachedWidth) return linesCache;
    cachedScrollHeight = sh;
    cachedWidth = w;
    var byTop = {};
    for (var i = 0; i < words.length; i++) {
      var r = words[i].getBoundingClientRect();
      var top = Math.round(r.top + window.scrollY);
      if (!byTop[top]) byTop[top] = [];
      byTop[top].push({ i: i, left: r.left + window.scrollX, centerX: r.left + window.scrollX + r.width / 2 });
    }
    var tops = Object.keys(byTop).map(Number).sort(function (a, b) { return a - b; });
    linesCache = tops.map(function (top) {
      var arr = byTop[top].sort(function (a, b) { return a.left - b.left; });
      return { top: top, words: arr };
    });
    return linesCache;
  }

  function findLine(idx) {
    var lines = ensureLines();
    for (var i = 0; i < lines.length; i++) {
      for (var j = 0; j < lines[i].words.length; j++) {
        if (lines[i].words[j].i === idx) return i;
      }
    }
    return -1;
  }

  function wordCenterX(idx) {
    var r = words[idx].getBoundingClientRect();
    return r.left + window.scrollX + r.width / 2;
  }

  function moveWord(dir) {
    setActive(activeIndex + dir);
    scrollActiveIntoView();
  }

  function moveSentence(dir) {
    var sid = sentenceIds[activeIndex];
    if (dir > 0) {
      for (var i = activeIndex + 1; i < words.length; i++) {
        if (sentenceIds[i] !== sid) {
          setActive(i);
          scrollActiveIntoView();
          return;
        }
      }
      setActive(words.length - 1);
      scrollActiveIntoView();
    } else {
      var target = sid - 1;
      if (target < 1) {
        setActive(0);
        scrollActiveIntoView();
        return;
      }
      for (var j = 0; j < words.length; j++) {
        if (sentenceIds[j] === target) {
          setActive(j);
          scrollActiveIntoView();
          return;
        }
      }
      setActive(0);
      scrollActiveIntoView();
    }
  }

  function moveBlock(dir) {
    var bid = blockIds[activeIndex];
    if (dir > 0) {
      for (var i = activeIndex + 1; i < words.length; i++) {
        if (blockIds[i] !== bid) {
          setActive(i);
          scrollActiveIntoView();
          return;
        }
      }
      setActive(words.length - 1);
      scrollActiveIntoView();
    } else {
      var target = bid - 1;
      if (target < 0) {
        setActive(0);
        scrollActiveIntoView();
        return;
      }
      for (var j = 0; j < words.length; j++) {
        if (blockIds[j] === target) {
          setActive(j);
          scrollActiveIntoView();
          return;
        }
      }
      setActive(0);
      scrollActiveIntoView();
    }
  }

  function moveLine(dir) {
    var li = findLine(activeIndex);
    if (li < 0) { moveWord(dir); return; }
    var target = (dir < 0) ? li - 1 : li + 1;
    var lines = ensureLines();
    if (target < 0 || target >= lines.length) { scrollActiveIntoView(); return; }
    var cx = wordCenterX(activeIndex);
    var arr = lines[target].words;
    var best = arr[0].i;
    var bestDist = Infinity;
    for (var i = 0; i < arr.length; i++) {
      var d = Math.abs(arr[i].centerX - cx);
      if (d < bestDist) { bestDist = d; best = arr[i].i; }
    }
    setActive(best);
    scrollActiveIntoView();
  }

  function moveLineStart() {
    var li = findLine(activeIndex);
    if (li < 0) return;
    setActive(ensureLines()[li].words[0].i);
    scrollActiveIntoView();
  }

  function moveLineEnd() {
    var li = findLine(activeIndex);
    if (li < 0) return;
    var arr = ensureLines()[li].words;
    setActive(arr[arr.length - 1].i);
    scrollActiveIntoView();
  }

  function moveDocEdge(dir) {
    setActive(dir < 0 ? 0 : words.length - 1);
    scrollActiveIntoView();
  }

  function markedNativeRange() {
    if (anchorIndex < 0) return null;
    var start = Math.min(anchorIndex, activeIndex);
    var end = Math.max(anchorIndex, activeIndex);
    var range = document.createRange();
    range.setStartBefore(words[start]);
    range.setEndAfter(words[end]);
    return range;
  }

  function clearMarked() {
    CSS.highlights.delete('ow-mark');
  }

  // Paints the range with the Custom Highlight API instead of a native
  // Selection: a live Selection would be painted by the theme's own
  // ::selection, and per-word spans leave the whitespace between them
  // unpainted. The highlight spans the whole range, gaps included; the
  // active word stays distinguishable through its .ow-active underline,
  // which draws over the highlight.
  function updateRange() {
    if (!marking || anchorIndex < 0) return;
    var range = markedNativeRange();
    if (range) CSS.highlights.set('ow-mark', new Highlight(range));
  }

  function clearSelection() {
    var sel = window.getSelection();
    if (sel) sel.removeAllRanges();
    clearMarked();
  }

  function selectedText() {
    var sel = window.getSelection();
    if (sel && !sel.isCollapsed && sel.toString().trim() !== '') return sel.toString();
    return null;
  }

  function showBadge(text, autoHideMs) {
    if (!badge) return;
    badge.textContent = text;
    applyBadgeColors();
    badge.classList.add('visible');
    if (badgeTimer) { clearTimeout(badgeTimer); badgeTimer = null; }
    if (autoHideMs) {
      badgeTimer = setTimeout(function () { badge.classList.remove('visible'); }, autoHideMs);
    }
  }

  function hideBadge() {
    if (badgeTimer) { clearTimeout(badgeTimer); badgeTimer = null; }
    if (badge) badge.classList.remove('visible');
  }

  function toggleMark() {
    if (marking) {
      marking = false;
      clearSelection();
      setActive(anchorIndex);
      anchorIndex = -1;
      hideBadge();
    } else {
      marking = true;
      anchorIndex = activeIndex;
      showBadge('SELECTING');
      updateRange();
    }
  }

  function flashActive() {
    if (!activeEl) return;
    activeEl.classList.add('ow-flash');
    setTimeout(function () { activeEl.classList.remove('ow-flash'); }, 500);
  }

  function copy() {
    // Priority: an active word-marking range, then a native mouse-drag
    // selection (if any), then the active word alone. The native Selection
    // itself is only ever created here, transiently, so execCommand('copy')
    // has something to act on, and is dropped again right after.
    var range = marking ? markedNativeRange() : null;
    var text = range ? range.toString() : selectedText();
    var targetWord = null;
    if (!range && (text === null || text === '')) {
      if (activeIndex >= 0 && words[activeIndex]) {
        text = words[activeIndex].textContent;
        targetWord = words[activeIndex];
        range = document.createRange();
        range.selectNodeContents(targetWord);
      } else {
        return;
      }
    }
    var sel = window.getSelection();
    if (range) {
      sel.removeAllRanges();
      sel.addRange(range);
    }
    var ok = false;
    try { ok = document.execCommand('copy'); } catch (e) { ok = false; }
    if (!ok && navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text);
    }
    sel.removeAllRanges();
    flashActive();
    showBadge('COPIED', 1200);
    if (marking) {
      marking = false;
      anchorIndex = -1;
      // keep the range highlight visible; native selection already dropped
    }
  }

  // Walks the same text nodes wrapWords() indexes (same EXCLUDED filter),
  // in document order, so a search match can span word boundaries or land
  // mid-word without needing a separate index from the one nav already
  // maintains. Returns parallel arrays: one text node per entry, and that
  // node's starting offset in the concatenation of all of them.
  function collectSearchNodes() {
    var walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT, null);
    var nodes = [];
    var starts = [];
    var pos = 0;
    var n;
    while ((n = walker.nextNode())) {
      if (!n.nodeValue) continue;
      var parent = n.parentElement;
      if (!parent) continue;
      if (parent.matches(EXCLUDED) || parent.closest(EXCLUDED)) continue;
      nodes.push(n);
      starts.push(pos);
      pos += n.nodeValue.length;
    }
    return { nodes: nodes, starts: starts };
  }

  // Binary search for the last start offset <= target; starts is sorted
  // ascending because collectSearchNodes walks in document order.
  function locateOffset(starts, target) {
    var lo = 0, hi = starts.length - 1, ans = 0;
    while (lo <= hi) {
      var mid = (lo + hi) >> 1;
      if (starts[mid] <= target) { ans = mid; lo = mid + 1; }
      else hi = mid - 1;
    }
    return ans;
  }

  function findMatches(query, caseSensitive) {
    var idx = collectSearchNodes();
    var full = idx.nodes.map(function (n) { return n.nodeValue; }).join('');
    var hay = caseSensitive ? full : full.toLowerCase();
    var needle = caseSensitive ? query : query.toLowerCase();
    var ranges = [];
    if (!needle) return ranges;
    var from = 0;
    while (true) {
      var found = hay.indexOf(needle, from);
      if (found < 0) break;
      var startNodeIdx = locateOffset(idx.starts, found);
      var endNodeIdx = locateOffset(idx.starts, found + needle.length - 1);
      var range = document.createRange();
      range.setStart(idx.nodes[startNodeIdx], found - idx.starts[startNodeIdx]);
      range.setEnd(idx.nodes[endNodeIdx], found + needle.length - idx.starts[endNodeIdx]);
      ranges.push(range);
      from = found + needle.length;
    }
    return ranges;
  }

  function paintSearch() {
    if (!searchMatches.length) {
      CSS.highlights.delete('ow-search');
      CSS.highlights.delete('ow-search-current');
      return;
    }
    CSS.highlights.set('ow-search', new Highlight(...searchMatches));
    CSS.highlights.set('ow-search-current', new Highlight(searchMatches[searchIndex]));
  }

  function scrollRangeIntoView(range) {
    var r = range.getBoundingClientRect();
    var margin = 40;
    if (r.top < margin) window.scrollBy(0, r.top - margin);
    else if (r.bottom > window.innerHeight - margin) window.scrollBy(0, r.bottom - (window.innerHeight - margin));
  }

  function searchResult() {
    return { count: searchMatches.length, current: searchMatches.length ? searchIndex + 1 : 0 };
  }

  function search(query, opts) {
    opts = opts || {};
    searchMatches = query ? findMatches(query, !!opts.caseSensitive) : [];
    searchIndex = searchMatches.length ? 0 : -1;
    paintSearch();
    if (searchIndex >= 0) scrollRangeIntoView(searchMatches[searchIndex]);
    return searchResult();
  }

  function searchStep(dir) {
    if (!searchMatches.length) return searchResult();
    searchIndex = (searchIndex + dir + searchMatches.length) % searchMatches.length;
    paintSearch();
    scrollRangeIntoView(searchMatches[searchIndex]);
    return searchResult();
  }

  function clearSearch() {
    CSS.highlights.delete('ow-search');
    CSS.highlights.delete('ow-search-current');
    searchMatches = [];
    searchIndex = -1;
  }

  // Prefill source for the search input: the word-marked range if active,
  // otherwise a native mouse-drag selection — same priority copy() uses.
  function selectionOrMarkedText() {
    var range = marking ? markedNativeRange() : null;
    if (range) return range.toString();
    return selectedText() || '';
  }

  function onKeydown(e) {
    if (!words.length) return;
    if (e.altKey) return;
    var key = e.key;
    var ctrl = e.ctrlKey;
    var handled = true;
    if (e.metaKey && (key === 'c' || key === 'C')) {
      copy();
    } else if (key === ' ' || key === 'Spacebar') {
      toggleMark();
    } else if (key === 'ArrowRight') {
      ctrl ? moveSentence(1) : moveWord(1);
    } else if (key === 'ArrowLeft') {
      ctrl ? moveSentence(-1) : moveWord(-1);
    } else if (key === 'ArrowDown') {
      ctrl ? moveBlock(1) : moveLine(1);
    } else if (key === 'ArrowUp') {
      ctrl ? moveBlock(-1) : moveLine(-1);
    } else if (key === 'Home') {
      ctrl ? moveDocEdge(-1) : moveLineStart();
    } else if (key === 'End') {
      ctrl ? moveDocEdge(1) : moveLineEnd();
    } else {
      handled = false;
    }
    if (handled) e.preventDefault();
  }

  function onMousedown(e) {
    if (e.button !== 0) return;
    if (!e.target || !e.target.closest) return;
    var el = e.target.closest('.ow-word');
    if (!el) return;
    var idx = words.indexOf(el);
    if (idx < 0 || idx === activeIndex) return;
    activeEl.classList.remove('ow-active');
    activeIndex = idx;
    activeEl = el;
    el.classList.add('ow-active');
    marking = false;
    anchorIndex = -1;
    hideBadge();
    clearSelection();
  }

  function ensureChrome() {
    if (!badge) {
      badge = document.createElement('div');
      badge.className = 'ow-nav-badge';
      document.body.appendChild(badge);
    }
  }

  function init() {
    words = [];
    blockIds = [];
    sentenceIds = [];
    activeIndex = -1;
    anchorIndex = -1;
    marking = false;
    activeEl = null;
    linesCache = null;
    cachedScrollHeight = -1;
    cachedWidth = -1;
    clearMarked();
    wrapWords();
    assignStructure();
    ensureChrome();
    applyHighlightColors();
    hideBadge();
    if (words.length > 0) setActive(0);
  }

  function serialize() {
    return {
      activeIndex: activeIndex,
      anchorIndex: anchorIndex,
      marking: marking
    };
  }

  function restore(state) {
    if (!state) return;
    if (typeof state.activeIndex === 'number' &&
        state.activeIndex >= 0 && state.activeIndex < words.length) {
      setActive(state.activeIndex);
    }
    if (state.marking && typeof state.anchorIndex === 'number' &&
        state.anchorIndex >= 0 && state.anchorIndex < words.length) {
      marking = true;
      anchorIndex = state.anchorIndex;
      showBadge('SELECTING');
      updateRange();
    }
  }

  document.addEventListener('keydown', onKeydown);
  document.addEventListener('mousedown', onMousedown);

  window.oMark = {
    init: init,
    serialize: serialize,
    restore: restore,
    copy: copy,
    search: search,
    searchNext: function () { return searchStep(1); },
    searchPrev: function () { return searchStep(-1); },
    clearSearch: clearSearch,
    selectionOrMarkedText: selectionOrMarkedText
  };

  init();
})();
