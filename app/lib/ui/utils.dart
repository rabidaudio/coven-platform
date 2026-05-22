import 'dart:collection';

import 'package:flutter/scheduler.dart';
import 'package:flutter/material.dart';

class SnackbarMessage {
  String message;
  Duration duration;
  SnackBarBehavior behavior;

  SnackbarMessage(
    this.message, {
    required this.duration,
    required this.behavior,
  });
}

mixin SnackbarState<T extends StatefulWidget> on State<T> {
  final Queue<SnackbarMessage> _messages = Queue();

  void pushSnackbar(
    String message, {
    Duration duration = const Duration(seconds: 2),
    SnackBarBehavior behavior = SnackBarBehavior.floating,
  }) {
    setState(() {
      _messages.add(
        SnackbarMessage(message, duration: duration, behavior: behavior),
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    showSnackbar(context);
    return Container();
  }

  void showSnackbar(BuildContext context) {
    if (_messages.isEmpty) return;

    SchedulerBinding.instance.addPostFrameCallback((_) {
      while (_messages.isNotEmpty) {
        final msg = _messages.removeFirst();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(msg.message),
            duration: msg.duration,
            behavior: msg.behavior,
          ),
        );
      }
    });
  }
}

enum NavigationCommand { push, pushReplacement }

class Navigation {
  final NavigationCommand cmd;
  final WidgetBuilder builder;

  Navigation(this.cmd, this.builder);
}

mixin RouterState<T extends StatefulWidget> on State<T> {
  final Queue<Navigation> _navigations = Queue();

  void pushNavigation(WidgetBuilder builder, {mode = NavigationCommand.push}) {
    setState(() {
      _navigations.add(Navigation(mode, builder));
    });
  }

  @override
  Widget build(BuildContext context) {
    navigate(context);
    return Container();
  }

  void navigate(BuildContext context) {
    if (_navigations.isEmpty) return;

    SchedulerBinding.instance.addPostFrameCallback((_) {
      while (_navigations.isNotEmpty) {
        final nav = _navigations.removeFirst();
        final route = MaterialPageRoute<void>(builder: nav.builder);

        switch (nav.cmd) {
          case NavigationCommand.push:
            Navigator.push(context, route);
          case NavigationCommand.pushReplacement:
            Navigator.pushReplacement(context, route);
        }
      }
    });
  }
}
