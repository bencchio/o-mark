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

    property real savedScrollY: 0
    property bool firstLoadComplete: false

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

    // Preserve scroll position across reloads (doc reload or theme/palette change).
    // On the initial load firstLoadComplete is false, so we skip the JS round-trip.
    onActiveHtmlChanged: {
        if (!activeHtml) return
        if (!firstLoadComplete) {
            webView.loadHtml(activeHtml, "o-mark://document")
            return
        }
        webView.runJavaScript("window.scrollY", function(y) {
            savedScrollY = y || 0
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
            if (savedScrollY > 0)
                webView.runJavaScript("window.scrollTo(0, " + savedScrollY + ")")
            var scripts; try { scripts = JSON.parse(postLoadScripts) } catch(e) { scripts = [] }
            for (var i = 0; i < scripts.length; i++)
                webView.runJavaScript(scripts[i])
        }

        onNavigationRequested: function(request) {
            var t = request.navigationType
            if (t !== WebEngineNavigationRequest.LinkClickedNavigation) return
            var url = request.url.toString()
            if (url.startsWith("http://") || url.startsWith("https://")) {
                request.reject()
                Qt.openUrlExternally(request.url)
                return
            }
            // Relative .md links resolve to o-mark://document/path — open in a new O'Mark instance.
            if (url.startsWith("o-mark://document/")) {
                var relPath
                try {
                    relPath = decodeURIComponent(url.slice("o-mark://document/".length).split("#")[0])
                } catch(e) {
                    request.reject()
                    return
                }
                if (relPath.split('/').some(function(p) { return p === '..'; })) {
                    request.reject()
                    return
                }
                if (relPath.toLowerCase().endsWith(".md"))
                    Qt.openUrlExternally("file://" + docDir + "/" + relPath)
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
