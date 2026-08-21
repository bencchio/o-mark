import QtQuick 2.15

QtObject {
    readonly property var _omarchy: { try { return JSON.parse(omarchyPaletteJson) } catch(e) { return {} } }

    readonly property color background: _omarchy.background || "#ffffff"
    readonly property color foreground: _omarchy.foreground || "#24292e"
    readonly property color accent:     _omarchy.accent     || "#0366d6"
    readonly property color surface:    _omarchy.surface    || "#f6f8fa"
    readonly property color border:     _omarchy.border     || "#dfe2e5"
}
