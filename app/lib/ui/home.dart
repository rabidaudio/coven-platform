import 'package:flutter/widget_previews.dart';
import 'package:flutter/material.dart';

import 'package:app/repos/nfc.dart';
import 'package:app/repos/token.dart';
import 'package:app/ui/login.dart';
import 'package:app/ui/splash.dart';
import 'package:app/ui/utils.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  _HomePageState createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> with RouterState {
  NFCState nfcState = NFCState.unknown;

  @override
  void initState() {
    super.initState();
    _checkLoggedIn();
    NFC.instance.state.addListener(_nfcStateListener);
  }

  @override
  void dispose() {
    NFC.instance.state.removeListener(_nfcStateListener);
    super.dispose();
  }

  void _nfcStateListener() {
    setState(() {
      nfcState = NFC.instance.state.value;
    });
  }

  Future<void> _checkLoggedIn() async {
    final hasToken = await TokenManager.hasToken();
    if (!hasToken) {
      pushNavigation(
        (context) => const SplashPage(),
        mode: NavigationCommand.pushReplacement,
      );
    }
  }

  Future<void> logout() async {
    await TokenManager.deleteToken();
    pushNavigation((context) => const LoginPage());
  }

  @override
  Widget build(BuildContext context) {
    super.build(context);
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
      body: Center(
        child: NFCView(
          nfcEnabled: true, // STOPSHIP
          state: nfcState,
        ),
      ),
    );
  }
}

class NFCView extends StatelessWidget {
  bool nfcEnabled;
  NFCState state;

  @Preview()
  NFCView({super.key, this.nfcEnabled = false, this.state = NFCState.unknown});

  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).colorScheme.secondary;
    var msg;
    if (!nfcEnabled) {
      msg = "Enable NFC";
    } else {
      switch (state) {
        case NFCState.connected:
        case NFCState.authenticating:
          msg = "Unlocking...";
        case NFCState.failedTokenExpired:
        case NFCState.failedOther:
          msg = "Unable to unlock";
        case NFCState.unlocked:
          msg = "Door Unlocked!";
        case NFCState.unknown:
        case NFCState.disconnected:
          msg = "Hold to door lock";
      }
    }

    return Container(
      decoration: BoxDecoration(
        border: BoxBorder.all(color: color, width: 10),
        borderRadius: BorderRadius.all(Radius.circular(5)),
      ),
      child: Column(
        mainAxisAlignment: .center,
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Icon(
              nfcEnabled
                  ? Icons.contactless_outlined
                  : Icons.sensors_off_outlined,
              color: color,
              size: 100,
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(bottom: 16),
            child: Text(
              msg,
              style: Theme.of(
                context,
              ).textTheme.displayMedium!.copyWith(color: color),
            ),
          ),
        ],
      ),
    );
  }
}
