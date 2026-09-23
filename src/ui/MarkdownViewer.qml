import QtQuick 2.15
import QtWebEngine 1.15

Item {
    id: viewer
    property var colors: null
    property int viewerThemeIndex: 0
    readonly property bool hasFocus: webView.activeFocus
    readonly property real zoomFactor: webView.zoomFactor

    function resetZoom() { webView.zoomFactor = 1.0 }

    readonly property var viewerThemes: { try { return JSON.parse(viewerThemesJson) } catch(e) { return [] } }
    readonly property var activeTheme: viewerThemes[viewerThemeIndex] || viewerThemes[0]
    readonly property string activeHtml: activeTheme ? activeTheme.html : ""
    readonly property color viewerBg: activeTheme ? activeTheme.bg : "#ffffff"

    function requestFocus() { webView.forceActiveFocus() }

    signal searchResultsChanged(int count, int current)

    function _applySearchResult(json) {
        var r = {}
        try { r = JSON.parse(json || "{}") } catch(e) {}
        searchResultsChanged(r.count || 0, r.current || 0)
    }

    function searchDocument(query, caseSensitive) {
        var q = JSON.stringify(query)
        webView.runJavaScript(
            "JSON.stringify(window.oMark.search(" + q + ", {caseSensitive: " + !!caseSensitive + "}))",
            _applySearchResult)
    }

    function searchNext() {
        webView.runJavaScript("JSON.stringify(window.oMark.searchNext())", _applySearchResult)
    }

    function searchPrev() {
        webView.runJavaScript("JSON.stringify(window.oMark.searchPrev())", _applySearchResult)
    }

    function clearSearch() {
        webView.runJavaScript("window.oMark.clearSearch()")
    }

    // Reads the current word-marked range or native selection to prefill the
    // search input; callback receives a plain string (possibly empty).
    function selectionOrMarkedText(callback) {
        webView.runJavaScript(
            "window.oMark ? window.oMark.selectionOrMarkedText() : ''",
            function(text) { callback(text || "") })
    }

    property real savedScrollY: 0
    property string savedNavState: ""
    property bool firstLoadComplete: false
    property bool _pdfCapture: false
    property string _pdfPath: ""

    signal pdfFinished(bool success)

    function exportPdf(path, html) {
        if (_pdfCapture || !path || !html) return
        _pdfCapture = true
        _pdfPath = path
        webView.loadHtml(html, "o-mark://document")
    }

    Shortcut {
        sequences: [StandardKey.ZoomIn, "Ctrl+="]
        onActivated: webView.zoomFactor = Math.min(2.0, Math.round((webView.zoomFactor + 0.1) * 10) / 10)
    }
    Shortcut {
        sequences: [StandardKey.ZoomOut]
        onActivated: webView.zoomFactor = Math.max(0.5, Math.round((webView.zoomFactor - 0.1) * 10) / 10)
    }
    Shortcut {
        sequence: "Ctrl+0"
        onActivated: webView.zoomFactor = 1.0
    }

    // Preserve scroll position and the navigation cursor across reloads (doc
    // reload or theme/palette change). On the initial load firstLoadComplete is
    // false, so we skip the JS round-trip.
    onActiveHtmlChanged: {
        if (_pdfCapture) return
        if (!activeHtml) return
        if (!firstLoadComplete) {
            var baseUrl = initialAnchor ? ("o-mark://document#" + initialAnchor) : "o-mark://document"
            webView.loadHtml(activeHtml, baseUrl)
            return
        }
        webView.runJavaScript(
            "JSON.stringify({y: window.scrollY, nav: window.oMark ? window.oMark.serialize() : null})",
            function(json) {
                var state = {}
                try { state = JSON.parse(json || "{}") } catch(e) {}
                savedScrollY = state.y || 0
                savedNavState = state.nav ? JSON.stringify(state.nav) : ""
                webView.loadHtml(activeHtml, "o-mark://document")
            })
    }

    WebEngineView {
        id: webView
        anchors.fill: parent
        backgroundColor: viewer.viewerBg
        settings.showScrollBars: showScrollbars

        onLoadingChanged: function(loadRequest) {
            if (loadRequest.status !== WebEngineView.LoadSucceededStatus) return
            firstLoadComplete = true
            var scripts; try { scripts = JSON.parse(postLoadScripts) } catch(e) { scripts = [] }
            for (var i = 0; i < scripts.length; i++)
                webView.runJavaScript(scripts[i])
            if (_pdfCapture) {
                webView.printToPdf(_pdfPath)
                return
            }
            if (savedScrollY > 0)
                webView.runJavaScript("window.scrollTo(0, " + savedScrollY + ")")
            // Restore the navigation cursor (and marking selection, if any)
            // after the post-load scripts rebuild the word list. Scroll wins
            // over cursor: the active word stays logical and auto-scroll brings
            // it into view on the next navigation move.
            if (savedNavState)
                webView.runJavaScript("window.oMark && window.oMark.restore(" + savedNavState + ")")
        }

        onPdfPrintingFinished: function(filePath, success) {
            _pdfCapture = false
            _pdfPath = ""
            if (viewer.activeHtml)
                webView.loadHtml(viewer.activeHtml, "o-mark://document")
            viewer.pdfFinished(success)
        }

        onNavigationRequested: function(request) {
            // Form submits go through the same scheme filter as link clicks: a
            // navigation is accepted unless rejected, so anything skipped here
            // bypasses the checks below. Other types (programmatic navigation,
            // including the loadHtml that renders the document) must pass.
            var t = request.navigationType
            if (t !== WebEngineNavigationRequest.LinkClickedNavigation
                && t !== WebEngineNavigationRequest.FormSubmittedNavigation) return
            var url = request.url.toString()
            if (url.startsWith("http://") || url.startsWith("https://")) {
                request.reject()
                Qt.openUrlExternally(request.url)
                return
            }
            // Relative .md links resolve to o-mark://document/path — open in a new
            // O'Mark instance, through the dedicated o-mark: scheme so the desktop
            // handoff (%u) carries the anchor instead of dropping it (see INSTALL.md).
            if (url.startsWith("o-mark://document/")) {
                var rest = url.slice("o-mark://document/".length)
                var hashIdx = rest.indexOf('#')
                var linkAnchor = hashIdx >= 0 ? rest.slice(hashIdx + 1) : ""
                var relPath
                try {
                    relPath = decodeURIComponent(hashIdx >= 0 ? rest.slice(0, hashIdx) : rest)
                } catch(e) {
                    request.reject()
                    return
                }
                if (relPath.split('/').some(function(p) { return p === '..'; })) {
                    request.reject()
                    return
                }
                if (relPath.toLowerCase().endsWith(".md")) {
                    var target = "o-mark://" + docDir + "/" + relPath
                    if (linkAnchor) target += "#" + linkAnchor
                    Qt.openUrlExternally(target)
                }
                request.reject()
                return
            }
            // Allow in-page anchor navigation (o-mark://document#anchor).
            if (url.startsWith("o-mark://document")) return
            request.reject()
        }

        onContextMenuRequested: function(request) { request.accepted = true }

        onNewWindowRequested: function(request) {
            var url = request.requestedUrl.toString()
            if (url.startsWith("http://") || url.startsWith("https://"))
                Qt.openUrlExternally(request.requestedUrl)
        }
    }
}
