package space.thecoven.android

/**
 * This class partially implements the protocol defined in ISO-7816-4.
 * TODO: support chaining, secure messaging
 */
object IDCard {

    class InvalidMessageException(message: String) : IllegalArgumentException(message)

    enum class LogicalChannel(val byte: UByte) {
        ZERO(0.toUByte()),
        ONE(1.toUByte()),
        TWO(2.toUByte()),
        THREE(3.toUByte())
    }

    // TODO: support BER-TLV encoding (odd values)
    enum class Command(i: Int) {
        DEACTIVATE_FILE(0x04),
        ERASE_RECORD(0x0C),
        ERASE_BINARY(0x0E),
        PERFORM_SCQL_OPERATION(0x10),
        PERFORM_TRANSACTION_OPERATION(0x12),
        PERFORM_USER_OPERATION(0x14),
        VERIFY(0x20),
        MANAGE_SECURITY_ENVIRONMENT(0x22),
        CHANGE_REFERENCE_DATA(0x24),
        DISABLE_VERIFICATION_REQUIREMENT(0x26),
        ENABLE_VERIFICATION_REQUIREMENT(0x28),
        PERFORM_SECURITY_OPERATION(0x2A),
        RESET_RETRY_COUNTER(0x2C),
        ACTIVATE_FILE(0x44),
        GENERATE_ASYMMETRIC_KEY_PAIR(0x46),
        MANAGE_CHANNEL(0x70),
        EXTERNAL_AUTHENTICATE(0x82),
        GET_CHALLENGE(0x84),
        GENERAL_AUTHENTICATE(0x86),
        INTERNAL_AUTHENTICATE(0x88),
        SEARCH_BINARY(0xA0),
        SEARCH_RECORD(0xA2),
        SELECT(0xA4),
        READ_BINARY(0xB0),
        READ_RECORD(0xB2),
        GET_RESPONSE(0xC0),
        ENVELOPE(0xC2),
        GET_DATA(0xCA),
        WRITE_BINARY(0xD0),
        WRITE_RECORD(0xD2),
        UPDATE_BINARY(0xD6),
        PUT_DATA(0xDA),
        UPDATE_RECORD(0xDC),
        CREATE_FILE(0xE0),
        APPEND_RECORD(0xE2),
        DELETE_FILE(0xE4),
        TERMINATE_DF(0xE6),
        TERMINATE_EF(0xE8),
        TERMINATE_CARD_USAGE(0xFE);

        val ins = i.toByte()
    }

    sealed class CommandAPDU {

        companion object {
            fun decode(message: ByteArray): CommandAPDU {
                if (message.isEmpty()) throw InvalidMessageException("Invalid message: not enough bytes")

                val cla = message[0]
                val clai = cla.toUByte().toInt()
                val interindustry = (clai.and(0b1000_0000).shr(7)) == 0

                if (!interindustry) return Proprietary(cla, message.copyOfRange(1, message.size))

                if (message.size < 4) throw InvalidMessageException("Invalid message: not enough bytes")
                val chaining = clai.and(0b0111_0000).shr(4).toByte()
                val secureMessaging = clai.and(0b0000_1100).shr(2).toByte()
                val channel = LogicalChannel.entries.find {
                    it.byte.toInt() == clai.and(0b0000_0011)
                }!!
                val ins = message[1]
                val command = Command.entries.find { it.ins == ins }
                    ?: throw InvalidMessageException("Unknown command: ${ins.toHexString()}")

                val p1 = message[2]
                val p2 = message[3]

                when (message.size) {
                    4 -> {
                        // lc and le must both be zero
                        val data = byteArrayOf()
                        val le = 0
                        return InterIndustry(
                            chaining = chaining,
                            secureMessaging = secureMessaging,
                            channel = channel,
                            command = command,
                            p1 = p1,
                            p2 = p2,
                            data = data,
                            maxResponseSize = le
                        )
                    }
                    5 -> {
                        // last byte must be Le
                        val data = byteArrayOf()
                        val le = message[4].toInt()
                        return InterIndustry(
                            chaining = chaining,
                            secureMessaging = secureMessaging,
                            channel = channel,
                            command = command,
                            p1 = p1,
                            p2 = p2,
                            data = data,
                            maxResponseSize = le
                        )
                    }
                    else -> {
                        val lc = if (message[4] == 0.toByte())
                            message[5].toUByte().toInt().shl(8) + message[6].toUByte().toInt()
                        else message[4].toUByte().toInt()
                        val dataStart = if (message[4] == 0.toByte()) 7 else 5
                        if (message.size < dataStart + lc)
                            throw InvalidMessageException("Invalid message: incomplete data")

                        val data = message.copyOfRange(dataStart, dataStart + lc)
                        val bytesRemain = message.size - dataStart - lc
                        val leIndex = dataStart + lc
                        val le = if (bytesRemain == 0) {
                            0
                        } else if (bytesRemain == 1) {
                            if (message[leIndex] == 0.toByte()) 256
                            else message[leIndex].toUByte().toInt()
                        } else if (bytesRemain == 2) {
                            if (message[leIndex] == 0.toByte() && message[leIndex + 1] == 0.toByte())
                                65536
                            else message[leIndex].toUByte().toInt()
                                .shl(8) + message[leIndex + 1].toUByte().toInt()
                        } else if (bytesRemain == 3) {
                            if (message[leIndex] != 0.toByte())
                                throw InvalidMessageException("Invalid message: unexpected Le")
                            if (message[leIndex + 1] == 0.toByte() && message[leIndex + 2] == 0.toByte())
                                65536
                            else message[leIndex + 1].toUByte().toInt()
                                .shl(8) + message[leIndex + 2].toUByte().toInt()
                        } else throw InvalidMessageException("Invalid message: unexpected Le")
                        return InterIndustry(
                            chaining = chaining,
                            secureMessaging = secureMessaging,
                            channel = channel,
                            command = command,
                            p1 = p1,
                            p2 = p2,
                            data = data,
                            maxResponseSize = le
                        )
                    }
                }
            }
        }

        data class Proprietary(val cla: Byte, val raw: ByteArray) : CommandAPDU() {
            override fun equals(other: Any?): Boolean {
                if (this === other) return true
                if (javaClass != other?.javaClass) return false

                other as Proprietary

                if (cla != other.cla) return false
                if (!raw.contentEquals(other.raw)) return false

                return true
            }

            override fun hashCode(): Int {
                var result = cla.toInt()
                result = 31 * result + raw.contentHashCode()
                return result
            }
        }

        data class InterIndustry(
            val chaining: Byte = 0,
            val secureMessaging: Byte = 0,
            val channel: LogicalChannel = LogicalChannel.ZERO,
            val command: Command,
            val p1: Byte = 0x00,
            val p2: Byte = 0x00,
            val data: ByteArray,
            val maxResponseSize: Int = 255,
        ) : CommandAPDU() {
            fun encode(): ByteArray {
                // lower 2 bits are logical channel. upper bits are for chaining and secure messaging
                val cla = (
                        chaining.toUByte().toInt().shl(4).and(0b0111_0000)
                                + secureMessaging.toUByte().toInt().shl(2).and(0b0000_1100)
                                + channel.byte.toInt().and(0b0000_0011)
                        ).toByte()
                val lc = encodeSize(data.size)
                val le = encodeSize(maxResponseSize)
                return byteArrayOf(cla, command.ins, p1, p2) + lc + data + le
            }

            private fun encodeSize(size: Int): ByteArray {
                return if (size == 0) {
                    // field should be absent
                    byteArrayOf()
                } else if (size <= 255) {
                    // short message
                    byteArrayOf(size.toByte())
                } else if (size > 65536) {
                    throw InvalidMessageException("size cannot exceed 65535 bytes")
                } else if (size == 65536) {
                    byteArrayOf(0, 0, 0)
                } else {
                    // long message
                    val a = size.shr(8).toByte()
                    val b = size.and(0xFF).toByte()
                    byteArrayOf(0, a, b)
                }
            }

            override fun equals(other: Any?): Boolean {
                if (this === other) return true
                if (javaClass != other?.javaClass) return false

                other as InterIndustry

                if (chaining != other.chaining) return false
                if (secureMessaging != other.secureMessaging) return false
                if (p1 != other.p1) return false
                if (p2 != other.p2) return false
                if (maxResponseSize != other.maxResponseSize) return false
                if (channel != other.channel) return false
                if (command != other.command) return false
                if (!data.contentEquals(other.data)) return false

                return true
            }

            override fun hashCode(): Int {
                var result = chaining.toInt()
                result = 31 * result + secureMessaging
                result = 31 * result + p1
                result = 31 * result + p2
                result = 31 * result + maxResponseSize
                result = 31 * result + channel.hashCode()
                result = 31 * result + command.hashCode()
                result = 31 * result + data.contentHashCode()
                return result
            }

        }
    }

    data class Status constructor(val sw1: UByte, val sw2: UByte) {

        constructor(a: Int, b: Int) : this(a.toUByte(), b.toUByte())

        companion object {
            val OK = Status(0x90, 0x00)
            val InvalidLength = Status(0x67, 0x00)

            val UnsupportedCLAFunctionLogicalChannel = Status(0x68, 0x81)
            val UnsupportedCLAFunctionSecureMessaging = Status(0x68, 0x82)
            val UnsupportedCLAFunctionCommandChanning = Status(0x68, 0x84)

            val UnsupportedCommand = Status(0x6D, 0x00)

            val FileOrApplicationNotFound = Status(0x6A, 0x82)
        }

        fun toResponse(data: ByteArray = byteArrayOf()) = ResponseAPDU(data, this)

        fun isOkay(): Boolean =
            (sw1 == 0x90.toUByte() && sw2 == 0x00.toUByte()) || sw1 == 0x61.toUByte()

        fun isWarning(): Boolean = sw1 == 0x62.toUByte() || sw1 == 0x63.toUByte()

        fun isError(): Boolean = !isOkay() && !isWarning()

        fun isInvalidLengthError(): Boolean = sw1 == 0x67.toUByte()

        fun functionIsUnsupported(): Boolean = sw1 == 0x68.toUByte()

        fun commandIsNotAllowed(): Boolean = sw1 == 0x69.toUByte()

        fun parametersAreInvalid(): Boolean = sw1 == 0x6A.toUByte() || sw1 == 0x6B.toUByte()

        fun maxLengthInvalid(): Boolean = sw1 == 0x6C.toUByte()

        fun instructionIsInvalidOrUnsupported(): Boolean = sw1 == 0x6D.toUByte()

        fun classIsNotSupported(): Boolean = sw1 == 0x6D.toUByte()

        fun remainingBytes(): Int {
            if (sw1 != 0x61.toUByte()) return 0
            return sw2.toInt()
        }

        fun asShort(): Short = ((sw1.toInt().shl(8)) + sw2.toInt()).toShort()

        override fun toString(): String {
            return "Status(${asShort().toHexString()})"
        }
    }

    data class ResponseAPDU(val data: ByteArray = byteArrayOf(), val status: Status = Status.OK) {
        companion object {
            fun parse(message: ByteArray): ResponseAPDU {
                if (message.size < 2) throw InvalidMessageException("Invalid message: not enough bytes")
                val dataSize = message.size - 2
                val data = message.copyOfRange(0, dataSize)
                val sw1 = message[dataSize].toUByte()
                val sw2 = message[dataSize + 1].toUByte()
                val status = Status(sw1, sw2)
                return ResponseAPDU(data, status)
            }
        }

        fun encode(): ByteArray {
            return data + byteArrayOf(status.sw1.toByte(), status.sw2.toByte()).also {
                if (it.size > 255) throw IllegalStateException("Message too big")
            }
        }

        override fun equals(other: Any?): Boolean {
            if (this === other) return true
            if (javaClass != other?.javaClass) return false

            other as ResponseAPDU

            if (!data.contentEquals(other.data)) return false
            if (status != other.status) return false

            return true
        }

        override fun hashCode(): Int {
            var result = data.contentHashCode()
            result = 31 * result + status.hashCode()
            return result
        }
    }
}
