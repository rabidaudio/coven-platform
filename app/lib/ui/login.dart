import 'package:app/nfc.dart';
import 'package:app/ui/splash.dart';
import 'package:flutter/material.dart';
import 'package:logging/logging.dart';

import '../api.dart';
import '../token.dart';
import '../prefs.dart';
import './main.dart';
import './utils.dart';
import 'styles.dart';

class LoginPage extends StatefulWidget {
  @override
  _LoginFormState createState() => _LoginFormState();
}

class _LoginFormState extends State<LoginPage> with SnackbarState, RouterState {
  final TextEditingController _email = TextEditingController();
  final TextEditingController _password = TextEditingController();

  @override
  void initState() {
    super.initState();
    _checkNfcState();
    _loadCachedEmail();
  }

  bool isLoggedIn = false;
  bool isLoading = false;

  Future<void> _checkNfcState() async {
    final currentState = await NFC.instance.refreshState();
    Logger.root.log(Level.INFO, "state at login: $currentState");
    if (currentState == NFCState.failedTokenExpired) {
      setState(() {
        pushSnackbar("Token invalid or expired. Log in and then try again");
      });
    }
  }

  Future<void> _loadCachedEmail() async {
    final email = await Prefs.getString("user.email");
    if (email != null && email != "") {
      // don't replace if the user already started typing
      if (_email.text.isEmpty) {
        _email.text = email;
      }
    }
  }

  Future<void> submit(String email, String password) async {
    setState(() {
      isLoading = true;
    });

    // cache email
    await Prefs.setString("user.email", email);

    try {
      await _login(email, password);
      setState(() {
        isLoading = false;
        isLoggedIn = true;
      });
      pushNavigation((context) {
        return MainPage();
      }, mode: NavigationCommand.pushReplacement);
    } on Exception catch (e) {
      if (e is ApiException && e.response.statusCode == 400) {
        setState(() {
          isLoggedIn = false;
          isLoading = false;
          pushSnackbar("Invalid username or password.");
        });
      } else {
        setState(() {
          isLoggedIn = false;
          isLoading = false;
          pushSnackbar("An error occurred, please try again later.");
        });
      }
    }
  }

  Future<void> _login(String email, String password) async {
    Map<String, dynamic> body = {"email": email, "password": password};
    final res = await Api.main().request("POST", "/session", body: body);
    final token = res
        .single(
          fromJson: (body) {
            return body["token"] as String;
          },
        )
        .data;
    await TokenManager.putToken(token);
  }

  @override
  Widget build(BuildContext context) {
    showSnackbar(context);
    navigate(context);

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text("Login"),
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: ConstrainedBox(
            constraints: BoxConstraints(maxWidth: 300),
            child: isLoading
                ? const Center(child: CircularProgressIndicator())
                : Column(
                    mainAxisAlignment: .center,
                    children: [
                      TextField(
                        autofocus: true,
                        decoration: Styles.textFieldDecoration.copyWith(
                          hintText: "email",
                        ),
                        controller: _email,
                      ),
                      TextField(
                        autofocus: true,
                        decoration: Styles.textFieldDecoration.copyWith(
                          hintText: "password",
                        ),
                        controller: _password,
                        obscureText: true,
                      ),
                      TextButton(
                        onPressed: () {
                          submit(_email.text, _password.text);
                        },
                        child: Text("Login"),
                      ),
                    ],
                  ),
          ),
        ),
      ),
    );
  }
}
