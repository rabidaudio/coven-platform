import 'package:shared_preferences/shared_preferences.dart';

class Prefs {
  static Future<String?> getString(String key) async {
    final prefs = await SharedPreferences.getInstance();
    return await prefs.getString(key);
  }

  static Future<void> setString(String key, String value) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(key, value);
  }
}
