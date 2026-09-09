import QtQuick 2.15

Window {
    id: root
    visible: true
    width: 800
    height: 600
    title: documentTitle ? "O'Mark — " + documentTitle : "o-mark"
    color: colors.background

    property int viewerThemeIndex: initialViewerThemeIndex
    property string toolbarPosition: toolbarPositionConfig
    property bool toolbarVisible: toolbarVisibleConfig
    property int docReloadTick: docReloadSignal

    // Badge colors derived from the active viewer theme background.
    readonly property real _viewerLum: 0.299 * viewer.viewerBg.r
                                     + 0.587 * viewer.viewerBg.g
                                     + 0.114 * viewer.viewerBg.b
    readonly property color _badgeBg: _viewerLum < 0.5
                                      ? Qt.lighter(viewer.viewerBg, 3.0)
                                      : Qt.darker(viewer.viewerBg, 1.5)
    readonly property color _badgeFg: _viewerLum < 0.5 ? "#d0d0d0" : "#303030"

    onDocReloadTickChanged: {
        if (docReloadTick > 0) reloadBadge.show()
    }

    Colors {
        id: colors
    }

    Shortcut { sequence: "Ctrl+Q"; onActivated: Qt.quit() }

    Component.onCompleted: viewer.requestFocus()

    // Esc toggles toolbar visibility. If the ComboBox popup is open, close it instead.
    Shortcut {
        sequence: "Escape"
        context: Qt.ApplicationShortcut
        onActivated: {
            if (toolbar.isPopupOpen) {
                toolbar.closePopup()
                return
            }
            root.toolbarVisible = !root.toolbarVisible
            if (root.toolbarVisible) toolbar.requestFocus()
            else viewer.requestFocus()
        }
    }

    Item {
        id: container
        anchors.fill: parent

        // All vertical positions are computed from parent.height and these two
        // values to avoid cross-sibling anchor dependencies that cause layout
        // glitches when toolbarPosition changes at runtime. The bar height comes
        // from the Toolbar itself, so it is defined in exactly one place.
        readonly property int _separator: 2
        readonly property int _chrome: toolbar.height + _separator

        MarkdownViewer {
            id: viewer
            anchors.left: parent.left
            anchors.right: parent.right
            y: root.toolbarPosition === "top" && root.toolbarVisible ? container._chrome : 0
            height: root.toolbarVisible ? parent.height - container._chrome : parent.height
            colors: colors
            viewerThemeIndex: root.viewerThemeIndex
        }

        Rectangle {
            id: separator
            anchors.left: parent.left
            anchors.right: parent.right
            // Top position: just below the toolbar. Bottom: just above it.
            y: root.toolbarPosition === "top" ? toolbar.height : parent.height - container._chrome
            height: container._separator
            visible: root.toolbarVisible
            color: toolbar.hasFocus ? colors.accent : colors.border
        }

        Toolbar {
            id: toolbar
            anchors.left: parent.left
            anchors.right: parent.right
            y: root.toolbarPosition === "top" ? 0 : parent.height - height
            visible: root.toolbarVisible
            colors: colors
            zoomFactor: viewer.zoomFactor
            onViewerThemeChanged: function(index) { root.viewerThemeIndex = index }
            onResetZoomRequested: viewer.resetZoom()
        }

        Rectangle {
            id: reloadBadge
            anchors.right: parent.right
            anchors.rightMargin: 12
            // Positioned 12px above the bottom edge of the viewer area.
            y: (root.toolbarPosition === "bottom" && root.toolbarVisible
                ? parent.height - container._chrome
                : parent.height) - height - 12
            width: reloadLabel.width + 16
            height: reloadLabel.height + 8
            radius: 4
            color: root._badgeBg
            border.width: 1
            border.color: root._badgeFg
            opacity: 0
            z: 10

            Text {
                id: reloadLabel
                anchors.centerIn: parent
                text: "Reloaded"
                color: root._badgeFg
                font.pixelSize: 12
            }

            function show() {
                opacity = 1
                hideTimer.restart()
            }

            Timer {
                id: hideTimer
                interval: 12000
                onTriggered: fadeOut.start()
            }

            NumberAnimation {
                id: fadeOut
                target: reloadBadge
                property: "opacity"
                to: 0
                duration: 800
            }
        }
    }
}
