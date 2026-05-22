//
//  Authenticator.swift
//  Runner
//
//  Created by Julien on 5/21/26.
//
import CryptoKit

enum AuthError : Error {
    case invalidMessage
    case invalidSignature
}

class Authenticator {
    
    let suite = HPKE.Ciphersuite(kem: HPKE.KEM.Curve25519_HKDF_SHA256, kdf: HPKE.KDF.HKDF_SHA256, aead: HPKE.AEAD.AES_GCM_128)
    
    func verify(challenge: Data) throws -> (Data, any HPKEDiffieHellmanPublicKey) {
        let challengeData = try TLVMessage.decode(data: challenge)
        guard let nonce = challengeData["nce"] else {
            throw AuthError.invalidMessage
        }
        guard let pub = challengeData["pub"] else {
            throw AuthError.invalidMessage
        }
        let pkR = try Curve25519.KeyAgreement.PublicKey(pub, kem: suite.kem)
        guard let sig = challengeData["sig"] else {
            throw AuthError.invalidMessage
        }
        
        let sigFieldSize = sig.count+4
        let unsignedMessageSize = challenge.count-sigFieldSize
        guard DoorSignPub.isValidSignature(sig, for: challenge.subdata(in: 0..<unsignedMessageSize)) else {
            throw AuthError.invalidSignature
        }
        return (nonce, pkR)
    }
    
    func encrypt(token: Data, nonce: Data, pubkey: any HPKEDiffieHellmanPublicKey) throws -> Data {
        var sender = try HPKE.Sender(recipientKey: pubkey, ciphersuite: suite, info: INFO)
        let enc = sender.encapsulatedKey
        let message = nonce + Data([UInt8(token.count)]) + token
        let ciphertext = try sender.seal(message)
        return enc + ciphertext
    }
}

enum TLVError : Error {
    case invalidField
    case invalidKey
}

struct TLVMessage {
    static func decode(data: Data) throws -> [String:Data] {
        var slice = data
        var fields: [String: Data] = [:]
        while true {
            if slice.isEmpty {
                return fields
            }
            guard slice.count >= 4 else {
                throw TLVError.invalidField
            }
            let tag = String(data: slice.subdata(in: 0..<3), encoding: .ascii)!
            let size = Int(slice[3])
            guard slice.count > size+4 else {
                throw TLVError.invalidField
            }
            fields[tag] = slice.subdata(in: 4..<(4+size))
            slice = slice.suffix(from: 4+size)
        }
    }

    static func encode(data: [String:Data]) throws -> Data {
        var result = Data()
        let keys = data.keys.sorted()
        for key in keys {
            let tag = key.data(using: .ascii)!
            guard tag.count == 3 else {
                throw TLVError.invalidKey
            }
            let size = data[key]!.count
            guard size < 256 else {
                throw TLVError.invalidField
            }
            let field = tag + Data([UInt8(size)]) + data[key]!
            result += field
        }
        return result
    }
}
