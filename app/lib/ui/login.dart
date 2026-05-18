import 'package:flutter/material.dart';

import '../api.dart';
import '../models/user.dart';
import 'styles.dart';

class LoginViewModel extends ChangeNotifier {
  User? user;
  Exception? error;
  bool isLoading = false;

  Future<void> load(String email, String password) async {
    isLoading = true;
    notifyListeners();

    try {
      user = await _login(email, password);
      error = null;
    } on Exception catch (e) {
      user = null;
      error = e;
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  Future<User> _login(String email, String password) async {
    Api api = Api(baseUrl: "http://localhost:8080");
    Map<String, dynamic> body = {"email": email, "password": password};
    ApiResponse<User> res = await api.post(
      "/session",
      body: body,
      fromJson: User.fromJson,
    );
    return res.data;
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
            child: ListenableBuilder(
              listenable: viewModel,
              builder: (context, _) => {
                return Text("Asdf");
                // if (viewModel.isLoading) {
                //   return const Center(child: CircularProgressIndicator());
                // } else {
                //   return Text("Asdf");
                // }
              },
            ),
          ),
        ),
      ),
    );
  }
}


// Column(
//               mainAxisAlignment: .center,
//               children: [
//                 TextField(
//                   autofocus: true,
//                   decoration: Styles.textFieldDecoration,
//                   controller: email,
//                 ),
//                 TextField(
//                   autofocus: true,
//                   decoration: Styles.textFieldDecoration,
//                   controller: password,
//                 ),
//                 TextButton(
//                   onPressed: () {
//                     viewModel.load(email.text, password.text);
//                   },
//                   child: Text("Login"),
//                 ),
//               ],
//             );
