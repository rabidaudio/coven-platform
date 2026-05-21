import 'package:app/nfc.dart';
import 'package:app/token.dart';
import 'package:app/ui/login.dart';
import 'package:app/ui/utils.dart';
import 'package:flutter/material.dart';

class MainPage extends StatefulWidget {
  @override
  _MainPageState createState() => _MainPageState();
}

class _MainPageState extends State<MainPage> with RouterState {
  Future<void> logout() async {
    await TokenManager.deleteToken();
    setState(() {
      pushNavigation((context) => LoginPage());
    });
  }

  @override
  Widget build(BuildContext context) {
    navigate(context);
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text("The Coven"),
        actions: [
          PopupMenuButton(
            onSelected: (value) {
              switch (value) {
                case "logout":
                  logout();
              }
            },
            itemBuilder: (context) => <PopupMenuEntry>[
              const PopupMenuItem(value: "logout", child: Text("Logout")),
            ],
          ),
        ],
      ),
      body: Center(child: _NFCView()),
    );
  }
}

class _NFCView extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return ValueListenableBuilder<NFCState>(
      valueListenable: NFC.instance.state,
      builder: (context, value, child) => Text("NFC State: $value"),
    );
  }
}
