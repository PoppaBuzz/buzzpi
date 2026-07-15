package com.jphat.buzzpi.domain.model

data class TerminalState(
    val sessionId: String = "",
    val lines: List<TerminalLine> = emptyList(),
    val cursorPosition: CursorPosition = CursorPosition(0, 0),
    val cursorVisible: Boolean = true,
    val dimensions: Dimensions = Dimensions(80, 24),
    val isConnected: Boolean = false,
    val error: String? = null
)

data class TerminalLine(
    val text: String = "",
    val segments: List<StyledSegment> = emptyList()
)

data class StyledSegment(
    val text: String,
    val fgColor: Int = 0xFFF1F5F9.toInt(),
    val bgColor: Int = 0x00000000.toInt(),
    val bold: Boolean = false,
    val italic: Boolean = false,
    val underline: Boolean = false
)

data class CursorPosition(
    val row: Int,
    val col: Int
)

data class Dimensions(
    val cols: Int,
    val rows: Int
)

data class TerminalInput(
    val data: ByteArray
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is TerminalInput) return false
        return data.contentEquals(other.data)
    }

    override fun hashCode(): Int = data.contentHashCode()
}

data class TerminalOutput(
    val sessionId: String,
    val output: ByteArray
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is TerminalOutput) return false
        return sessionId == other.sessionId && output.contentEquals(other.output)
    }

    override fun hashCode(): Int = 31 * sessionId.hashCode() + output.contentHashCode()
}
