package space.thecoven.android

import java.nio.ByteBuffer

/**
 * Implementation of Type-Length-Value (TLV) encoding/decoding with 3-byte tags and 1-byte lengths.
 */
class TLV(
    val wrapped: MutableMap<String, ByteArray> = mutableMapOf()
) : Map<String, ByteArray> by wrapped {

    companion object {
        fun decode(data: ByteArray): TLV {
            var slice = data
            val map = mutableMapOf<String, ByteArray>()
            while (true) {
                if (slice.isEmpty()) return TLV(map)

                if (slice.size < 4) throw IllegalArgumentException("Invalid length of tag")
                val tag = slice.sliceArray(0..<3).toString(Charsets.US_ASCII)
                val size = slice[3].toUByte().toInt()
                if (slice.size < size+4) throw IllegalArgumentException("Invalid length of content")
                map[tag] = slice.sliceArray(4..<4+size)
                slice = slice.sliceArray(4+size..<slice.size)
            }
        }
    }

    fun setUint64(key: String, value: ULong) {
        wrapped[key] = ByteBuffer.allocate(8).apply {
            putLong(value.toLong())
        }.array()
    }

    fun getUint64(key: String): ULong {
        return ByteBuffer.wrap(wrapped[key]!!).getLong().toULong()
    }

    fun validate() {
        for (key in keys()) {
            if (key.length != 3) throw IllegalStateException("Key must have length 3: $key")
            val size = wrapped[key]!!.size
            if (size >= 256) throw IllegalStateException("Value of key $key too long: $size")
        }
    }

    fun encode(): ByteArray {
        validate()
        val buf = ByteBuffer.allocate(256)
        for (key in keys()) {
            buf.put(key.toByteArray(Charsets.US_ASCII))
            buf.put(wrapped[key]!!.size.toUByte().toByte())
            buf.put(wrapped[key]!!)
        }
        return buf.array()
    }

    private fun keys() = wrapped.keys.sorted()
}