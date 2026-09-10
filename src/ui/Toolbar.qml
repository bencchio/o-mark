import QtQuick 2.15
import QtQuick.Controls.Basic 2.15
import QtQuick.Layouts 1.15

Rectangle {
    id: toolbarRoot
    property var colors: null
    property real zoomFactor: 1.0
    // 0 = search (real stop only while searchActive), 1 = ↺ reset, 2 = ComboBox
    property int toolbarActiveIndex: 2

    property bool searchActive: false
    property int searchCount: 0
    property int searchCurrent: 0
    property bool searchCaseSensitive: false

    signal viewerThemeChanged(int index)
    signal resetZoomRequested()
    signal searchRequested(string query, bool caseSensitive)
    signal searchNextRequested()
    signal searchPrevRequested()
    signal searchClosed()

    function _fileName(path) {
        var parts = (path || "").split("/")
        return parts[parts.length - 1]
    }

    // Opens the search block, prefills it from the word-marked range or
    // native selection (empty string if neither), and runs a search right
    // away when there is something to prefill with.
    function openSearch(prefill) {
        searchActive = true
        searchField.text = prefill || ""
        setToolbarElement(0)
        if (prefill) toolbarRoot.searchRequested(prefill, searchCaseSensitive)
    }

    // Closes only the search block — the toolbar itself stays as it was.
    // searchCaseSensitive is deliberately not reset: it persists for the
    // session, across separate searches.
    function closeSearch() {
        if (!searchActive) return
        searchActive = false
        searchField.text = ""
        searchCount = 0
        searchCurrent = 0
        toolbarRoot.searchClosed()
    }

    function setSearchResult(count, current) {
        searchCount = count
        searchCurrent = current
    }

    // Shared by the Aa click and its keyboard paths (Enter/Space with
    // keyboard focus, and the "Escape,A" shortcut in main.qml) so all three
    // stay in sync.
    function toggleCaseSensitive() {
        searchCaseSensitive = !searchCaseSensitive
        toolbarRoot.searchRequested(searchField.text, searchCaseSensitive)
    }

    height: 36
    color: colors ? colors.background : "#f6f8fa"

    // true when any toolbar element has active focus
    readonly property bool hasFocus: comboBox.activeFocus || searchField.activeFocus || activeFocus

    readonly property bool isPopupOpen: comboBox.popup.visible
    function closePopup() { comboBox.popup.close() }

    // Restore focus to the last active element (persists across Esc toggles).
    function requestFocus() { setToolbarElement(toolbarActiveIndex) }

    function setToolbarElement(idx) {
        toolbarActiveIndex = idx
        if (idx === 0) searchField.forceActiveFocus()
        else if (idx === 2) comboBox.forceActiveFocus()
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
            text: "O'Mark — " + (toolbarRoot.searchActive ? toolbarRoot._fileName(documentTitle) : documentTitle)
            font.family: omarchyFont
            font.bold: true
            font.pixelSize: 12
            color: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
            elide: Text.ElideMiddle
            Layout.maximumWidth: 260
            Layout.alignment: Qt.AlignVCenter
        }

        Row {
            id: searchBlock
            visible: toolbarRoot.searchActive
            spacing: 6
            Layout.alignment: Qt.AlignVCenter

            Text {
                id: caseToggle
                text: "Aa"
                font.family: omarchyFont
                font.pixelSize: 12
                font.bold: toolbarRoot.searchCaseSensitive
                anchors.verticalCenter: parent.verticalCenter
                // Accent + underline when keyboard-focused; accent when active or hovered.
                readonly property bool keyboardFocused: caseToggle.activeFocus
                color: toolbarRoot.searchCaseSensitive || keyboardFocused || caseToggleArea.containsMouse
                       ? (toolbarRoot.colors ? toolbarRoot.colors.accent : "#0366d6")
                       : (toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e")
                font.underline: keyboardFocused

                Keys.onPressed: function(event) {
                    if (event.key === Qt.Key_Tab || event.key === Qt.Key_Backtab) {
                        searchField.forceActiveFocus()
                        event.accepted = true
                    } else if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter ||
                               event.key === Qt.Key_Space) {
                        toolbarRoot.toggleCaseSensitive()
                        event.accepted = true
                    }
                }

                MouseArea {
                    id: caseToggleArea
                    anchors.fill: parent
                    anchors.margins: -4
                    hoverEnabled: true
                    cursorShape: Qt.PointingHandCursor
                    onClicked: toolbarRoot.toggleCaseSensitive()
                }
            }

            TextField {
                id: searchField
                implicitWidth: 280
                anchors.verticalCenter: parent.verticalCenter
                font.family: omarchyFont
                font.pixelSize: 11
                placeholderText: qsTr("Buscar…")
                selectByMouse: true
                color: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
                background: Rectangle {
                    color: toolbarRoot.colors ? toolbarRoot.colors.surface : "#f6f8fa"
                    border.color: searchField.activeFocus
                                  ? (toolbarRoot.colors ? toolbarRoot.colors.accent : "#0366d6")
                                  : (toolbarRoot.colors ? toolbarRoot.colors.border : "#dfe2e5")
                    border.width: searchField.activeFocus ? 2 : 1
                    radius: 3
                }
                onTextChanged: toolbarRoot.searchRequested(text, toolbarRoot.searchCaseSensitive)
                Keys.onPressed: function(event) {
                    if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter) {
                        if (event.modifiers & Qt.ShiftModifier) toolbarRoot.searchPrevRequested()
                        else toolbarRoot.searchNextRequested()
                        event.accepted = true
                    } else if (event.key === Qt.Key_Tab || event.key === Qt.Key_Backtab) {
                        caseToggle.forceActiveFocus()
                        event.accepted = true
                    }
                }
            }

            Text {
                text: toolbarRoot.searchCount > 0
                      ? (toolbarRoot.searchCurrent + " / " + toolbarRoot.searchCount)
                      : (searchField.text.length > 0 ? qsTr("0 resultados") : "")
                color: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
                font.family: omarchyFont
                font.pixelSize: 11
                anchors.verticalCenter: parent.verticalCenter
            }
        }

        Item { Layout.fillWidth: true }

        Text {
            text: Math.round(toolbarRoot.zoomFactor * 100) + "%"
            color: toolbarRoot.colors ? toolbarRoot.colors.foreground : "#24292e"
            font.family: omarchyFont
            font.pixelSize: 11
            Layout.alignment: Qt.AlignVCenter
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
            Layout.alignment: Qt.AlignVCenter

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
            Layout.alignment: Qt.AlignVCenter

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
