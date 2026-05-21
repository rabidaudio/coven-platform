import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

enum NFCState {
  unknown,
  connected,
  authenticating,
  unlocked,
  failedTokenExpired,
  failedOther,
  disconnected,
}

class NFC {
  static final instance = NFC();

  // https://docs.flutter.dev/platform-integration/platform-channels
  final _interface = MethodChannel("space.thecoven/nfc");

  final state = ValueNotifier<NFCState>(NFCState.unknown);

  NFCState _fromString(String stateStr) {
    return NFCState.values.firstWhere((v) => v.name == stateStr);
  }

  NFC() {
    _interface.setMethodCallHandler((call) async {
      switch (call.method) {
        case "onNFCState":
          final stateStr = (call.arguments as String);
          state.value = _fromString(stateStr);
      }
    });
    // async load the current state
    _getNFCState().then((v) => state.value = v);
  }

  Future<NFCState> _getNFCState() async {
    if (Platform.isAndroid) {
      final stateStr = await _interface.invokeMethod<String>("getNFCState");
      return _fromString(stateStr!);
    } else {
      throw StateError("Unsupported platform");
    }
  }
}
