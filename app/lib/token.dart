import 'package:flutter_keychain/flutter_keychain.dart';

class TokenManager {
  static const String key = "user_token";

  static Future<bool> hasToken() async {
    var token = await FlutterKeychain.get(key: key);
    return token != null;
  }

  static Future<void> putToken(String token) async {
    await FlutterKeychain.put(key: key, value: token);
  }

  static Future<void> deleteToken() async {
    await FlutterKeychain.remove(key: key);
  }
}
