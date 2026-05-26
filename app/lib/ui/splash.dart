import 'dart:async';

import 'package:app/ui/utils.dart';
import 'package:flutter/material.dart';

import 'package:app/repos/token.dart';
import 'package:app/ui/home.dart';
import 'package:app/ui/login.dart';

class SplashPage extends StatefulWidget {
  const SplashPage({super.key});

  @override
  _SplashPageState createState() => _SplashPageState();
}

class _SplashPageState extends State<SplashPage> with RouterState {
  @override
  void initState() {
    super.initState();
    _start();
  }

  Future<void> _start() async {
    final hasToken = await TokenManager.hasToken();
    if (hasToken) {
      pushNavigation(
        (context) => const HomePage(),
        mode: NavigationCommand.pushReplacement,
      );
    } else {
      pushNavigation(
        (context) => const LoginPage(),
        mode: NavigationCommand.pushReplacement,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    super.build(context);
    return Scaffold(
      body: Center(
        // TODO: replace with logo
        child: Text("The Coven"),
      ),
    );
  }
}
