import 'package:app/ui/components/error_snackbar.dart';
import 'package:logging/logging.dart';
import 'package:flutter/material.dart';
import 'package:flutter_keychain/flutter_keychain.dart';

import '../api.dart';
import 'styles.dart';

class LoginViewModel extends ChangeNotifier with ErrorViewModel {
  bool isLoggedIn = false;
  bool isLoading = false;

  Future<void> submit(String email, String password) async {
    var existingToken = await FlutterKeychain.get(key: "user_token");
    Logger.root.log(Level.INFO, "token: ${existingToken}");

    isLoading = true;
    notifyListeners();

    try {
      await _login(email, password);
      isLoggedIn = true;
    } on Exception catch (e) {
      pushError(e);
      isLoggedIn = false;
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  Future<void> _login(String email, String password) async {
    Map<String, dynamic> body = {"email": email, "password": password};
    ApiResponse<String> res = await Api.main().post(
      "/session",
      body: body,
      fromJson: (body) {
        return body["token"] as String;
      },
    );
    await FlutterKeychain.put(key: "user_token", value: res.data);
  }
}

class LoginPage extends StatelessWidget {
  final LoginViewModel viewModel = LoginViewModel();

  final TextEditingController email = TextEditingController();
  final TextEditingController password = TextEditingController();

  @override
  Widget build(BuildContext context) {
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
            child: SnackbarError(
              vm: viewModel,
              child: ListenableBuilder(
                listenable: viewModel,
                builder: (context, _) {
                  if (viewModel.isLoading) {
                    return const Center(child: CircularProgressIndicator());
                  } else {
                    return Column(
                      mainAxisAlignment: .center,
                      children: [
                        TextField(
                          autofocus: true,
                          decoration: Styles.textFieldDecoration.copyWith(
                            hintText: "email",
                          ),
                          controller: email,
                        ),
                        TextField(
                          autofocus: true,
                          decoration: Styles.textFieldDecoration.copyWith(
                            hintText: "password",
                          ),
                          controller: password,
                          obscureText: true,
                        ),
                        TextButton(
                          onPressed: () {
                            viewModel.submit(email.text, password.text);
                          },
                          child: Text("Login"),
                        ),
                      ],
                    );
                  }
                },
              ),
            ),
          ),
        ),
      ),
    );
  }
}
