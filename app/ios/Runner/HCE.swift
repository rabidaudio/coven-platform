//
//  HCE.swift
//  Runner
//
//  Created by Julien on 5/21/26.
//
import CoreNFC

enum HCEError: Error {
    case Unsupported
    case Unavailable
}

enum NFCState : String {
    case unknown
    case connected
    case authenticating
    case unlocked
    case failedTokenExpired
    case failedOther
    case disconnected
}

// https://developer.apple.com/documentation/corenfc/cardsession/
class HCE {
    var state: NFCState = .unknown
    var channel: FlutterMethodChannel
    
    init(engineBridge: FlutterImplicitEngineBridge) {
        channel = FlutterMethodChannel(name: "space.thecoven/nfc", binaryMessenger: engineBridge.applicationRegistrar.messenger())
        
        channel.setMethodCallHandler({
            [weak self] (call: FlutterMethodCall, result: FlutterResult) -> Void in
            guard let self = self else {
                return
            }
            switch (call.method) {
            case "getNFCState":
                result(String(self.state.rawValue))
            default:
                result(FlutterMethodNotImplemented)
            }
        })
        
    }
    
    func setState(_ state: NFCState) {
        self.state = state
        channel.invokeMethod("onNFCState", arguments: state.rawValue)
    }
    
    // entitlements:
//    com.apple.developer.nfc.hce = true
//    com.apple.developer.nfc.hce.iso7816.select-identifier-prefixes ["ff746865636f76656e2e7370616365ff"]
    func unlock() async throws {
        guard NFCReaderSession.readingAvailable, CardSession.isSupported, await CardSession.isEligible else {
            throw HCEError.Unsupported
        }
        // Hold a presentment intent assertion reference to prevent the
        // default contactless app from launching. In a real app, monitor
        // presentmentIntent.isValid to ensure the assertion remains active.
        var presentmentIntent: NFCPresentmentIntentAssertion?

        let cardSession: CardSession
        do {
            presentmentIntent = try await NFCPresentmentIntentAssertion.acquire()
            cardSession = try await CardSession()

        } catch {
            throw HCEError.Unavailable
        }
        defer {
            cardSession.invalidate()
            presentmentIntent = nil /// Release presentment intent assertion.
        }

        // Iterate over events as the card session produces them.
       for try await event in cardSession.eventStream {
           switch event {
           case .sessionStarted:
               cardSession.alertMessage = String(localized: "Communicating with card reader.")
               setState(.connected)
               
           case .readerDetected:
               /// Start card emulation on first detection of an external reader.
               try await cardSession.startEmulation()
               
           case .readerDeselected:
               /// Stop emulation on first notification of RF link loss.
               await cardSession.stopEmulation(status: .success)
               setState(.disconnected)
               
           case .received(let cardAPDU):
               do {
                   /// Call handler to process received input and produce a response.
                   let responseAPDU = processAPDU(cardAPDU.payload)
                   
                   try await cardAPDU.respond(response: responseAPDU)
               } catch {
                   /// Handle the error from respond(response:). If the error is
                   /// CardSession.Error.transmissionError, then retry by calling
                   /// CardSession.APDU.respond(response:) again.
               }
               
           case .sessionInvalidated(reason: _):
               cardSession.alertMessage = String(localized: "Ending communication with card reader.")
               /// Handle the reason for session invalidation.
               
               break
           default:
               continue
           }
           guard presentmentIntent!.isValid else {
               if await cardSession.isEmulationInProgress {
                   await cardSession.stopEmulation(status: .failure)
               }
               setState(.failedOther)
               return
           }
       }
    }
    
    func processAPDU(_ cadpu: Data) -> Data {
        return Data() // TODO: implement algorithm
    }
}
