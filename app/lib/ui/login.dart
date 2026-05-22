import 'package:flutter/material.dart';
import 'package:logging/logging.dart';

import 'package:app/repos/api.dart';
import 'package:app/repos/token.dart';
import 'package:app/repos/prefs.dart';
import 'package:app/repos/nfc.dart';
import 'package:app/ui/home.dart';
import 'package:app/ui/utils.dart';
import 'package:app/ui/styles.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

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
      pushSnackbar("Token invalid or expired. Log in and then try again");
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
      pushNavigation(
        (context) => const HomePage(),
        mode: NavigationCommand.pushReplacement,
      );
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
    super.build(context);
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
