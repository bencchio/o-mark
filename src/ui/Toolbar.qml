import QtQuick 2.15
import QtQuick.Controls.Basic 2.15
import QtQuick.Layouts 1.15

Rectangle {
    id: toolbarRoot
    property var colors: null
    property real zoomFactor: 1.0
    // 0 = search (post-1.0, not yet interactive), 1 = ↺ reset, 2 = ComboBox
    property int toolbarActiveIndex: 2

    signal viewerThemeChanged(int index)
    signal resetZoomRequested()

    height: 36
    color: colors ? colors.background : "#f6f8fa"

    // true when any toolbar element has active focus
    readonly property bool hasFocus: comboBox.activeFocus || activeFocus

    readonly property bool isPopupOpen: comboBox.popup.visible
    function closePopup() { comboBox.popup.close() }

    // Restore focus to the last active element (persists across Esc toggles).
    function requestFocus() { setToolbarElement(toolbarActiveIndex) }

    function setToolbarElement(idx) {
        toolbarActiveIndex = idx
        if (idx === 2) comboBox.forceActiveFocus()
        else toolbarRoot.forceActiveFocus()
    }

    // T → focus theme ComboBox, Z → focus zoom reset ↺.
    // ← / → navigate between toolbar elements; Enter/Space activate ↺.
    // Events from ComboBox propagate here when ComboBox doesn't handle them.
    Keys.onPressed: function(event) {
        if (event.key === Qt.Key_T) {
            setToolbarElement(2)
            event.accepted = true
        } else if (event.key === Qt.Key_Z) {
            setToolbarElement(1)
            event.accepted = true
        } else if (event.key === Qt.Key_Left || event.key === Qt.Key_Right) {
            var next = toolbarActiveIndex + (event.key === Qt.Key_Right ? 1 : -1)
            if (next < 1) next = 2
            if (next > 2) next = 1
            setToolbarElement(next)
            event.accepted = true
        } else if (toolbarActiveIndex === 1 &&
                   (event.key === Qt.Key_Return || event.key === Qt.Key_Enter ||
                    event.key === Qt.Key_Space)) {
            toolbarRoot.resetZoomRequested()
            event.accepted = true
        }
    }

    RowLayout {
        anchors.verticalCenter: parent.verticalCenter
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.leftMargin: 12
        anchors.rightMargin: 12
        spacing: 8

        Label {
            text: "O'Mark — " + documentTitle
            font.family: omarchyFont
            font.bold: true
            font.pixelSize: 12
            color: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
        }

        // Placeholder for future Ctrl+F search bar (post-1.0).
        Item { implicitWidth: 0 }

        Item { Layout.fillWidth: true }

        Text {
            text: Math.round(toolbarRoot.zoomFactor * 100) + "%"
            color: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
            font.family: omarchyFont
            font.pixelSize: 11
        }

        Text {
            id: resetButton
            text: "↺"
            // Accent + underline when keyboard-focused; accent on hover; foreground otherwise.
            readonly property bool keyboardFocused: toolbarRoot.toolbarActiveIndex === 1 && toolbarRoot.activeFocus
            color: keyboardFocused || resetArea.containsMouse
                   ? (toolbarRoot.colors ? toolbarRoot.colors.accent : "#0366d6")
                   : (toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e")
            font.pixelSize: 14
            font.family: omarchyFont
            font.underline: keyboardFocused

            MouseArea {
                id: resetArea
                anchors.fill: parent
                anchors.margins: -4
                hoverEnabled: true
                cursorShape: Qt.PointingHandCursor
                onClicked: toolbarRoot.resetZoomRequested()
            }
        }

        ComboBox {
            id: comboBox
            focusPolicy: Qt.NoFocus
            model: { try { return JSON.parse(viewerThemeLabelsJson) } catch(e) { return [] } }
            Component.onCompleted: currentIndex = initialViewerThemeIndex
            onCurrentIndexChanged: toolbarRoot.viewerThemeChanged(currentIndex)
            font.family: omarchyFont
            font.pixelSize: 11
            implicitHeight: 24
            Layout.minimumWidth: 120

            background: Rectangle {
                color: toolbarRoot.colors ? toolbarRoot.colors.surface : "#f6f8fa"
                border.color: comboBox.activeFocus
                              ? (toolbarRoot.colors ? toolbarRoot.colors.accent : "#0366d6")
                              : (toolbarRoot.colors ? toolbarRoot.colors.border : "#dfe2e5")
                border.width: comboBox.activeFocus ? 2 : 1
                radius: 3
            }

            contentItem: Text {
                leftPadding: 8
                rightPadding: comboBox.indicator.width + 6
                text: comboBox.displayText
                font: comboBox.font
                color: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
                verticalAlignment: Text.AlignVCenter
                elide: Text.ElideRight
            }

            indicator: Text {
                x: parent.width - width - 6
                y: (parent.height - height) / 2
                text: "▾"
                font.pixelSize: 10
                color: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
            }

            delegate: ItemDelegate {
                id: itemDel
                width: Math.max(120, comboBox.width)
                height: 26
                leftPadding: 8
                rightPadding: 8
                text: modelData
                highlighted: comboBox.highlightedIndex === index
                palette.text: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
                palette.highlightedText: "#ffffff"
                font: comboBox.font

                background: Rectangle {
                    color: itemDel.highlighted
                           ? (toolbarRoot.colors ? toolbarRoot.colors.accent : "#0366d6")
                           : "transparent"
                    topLeftRadius:     index === 0                  ? 3 : 0
                    topRightRadius:    index === 0                  ? 3 : 0
                    bottomLeftRadius:  index === comboBox.count - 1 ? 3 : 0
                    bottomRightRadius: index === comboBox.count - 1 ? 3 : 0
                }
            }

            popup: Popup {
                y: comboBox.height + 4
                width: Math.max(120, comboBox.width)
                padding: 0

                background: Rectangle {
                    color: toolbarRoot.colors ? toolbarRoot.colors.surface : "#f6f8fa"
                    border.color: toolbarRoot.colors ? toolbarRoot.colors.border : "#dfe2e5"
                    border.width: 1
                    radius: 3
                }

                contentItem: ListView {
                    clip: true
                    implicitHeight: contentHeight
                    model: comboBox.popup.visible ? comboBox.delegateModel : null
                    currentIndex: comboBox.highlightedIndex
                    highlightMoveDuration: 0
                }
            }
        }
    }
}
