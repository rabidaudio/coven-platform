import 'dart:async';

import 'package:app/ui/main.dart';
import 'package:flutter/scheduler.dart';
import 'package:flutter/material.dart';

import '../token.dart';
import './login.dart';

enum Destination { login, main }

class SplashViewModel extends ChangeNotifier {
  final StreamController<Destination> _streamController = StreamController();

  SplashViewModel() {
    _start();
  }

  Stream<Destination> navDestinations() {
    return _streamController.stream;
  }

  Future<void> _start() async {
    final hasToken = await TokenManager.hasToken();
    if (hasToken) {
      _streamController.add(Destination.main);
    } else {
      _streamController.add(Destination.login);
    }
  }
}

class SplashPage extends StatelessWidget {
  final _vm = SplashViewModel();

  @override
  Widget build(BuildContext context) {
    return StreamBuilder(
      stream: _vm.navDestinations(),
      builder: (context, snapshot) {
        if (snapshot.hasData) {
          SchedulerBinding.instance.addPostFrameCallback((_) {
            Navigator.pushReplacement(
              context,
              MaterialPageRoute<void>(
                builder: (context) {
                  switch (snapshot.data!) {
                    case Destination.login:
                      return LoginPage();
                    case Destination.main:
                      return MainPage();
                  }
                },
              ),
            );
          });
        }

        return Scaffold(
          body: Center(
            // TODO: replace with logo
            child: Text("The Coven"),
          ),
        );
      },
    );
  }
}
