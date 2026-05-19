import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:logging/logging.dart';

mixin ErrorViewModel on ChangeNotifier {
  final StreamController<Exception> _streamController = StreamController();

  Stream<Exception> errors() {
    return _streamController.stream;
  }

  void pushError(Exception err) {
    _streamController.add(err);
    notifyListeners();
  }
}

class SnackbarError extends StatelessWidget {
  final ErrorViewModel vm;
  final String message;
  final Widget? child;

  const SnackbarError({
    required this.vm,
    this.message = 'An error occurred, please try again later',
    this.child,
  });

  @override
  Widget build(BuildContext context) {
    return StreamBuilder(
      stream: vm.errors(),
      builder: (context, snapshot) {
        if (snapshot.hasData) {
          Logger.root.log(Level.WARNING, snapshot.data!);
          SchedulerBinding.instance.addPostFrameCallback((_) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(message),
                duration: const Duration(seconds: 2),
                behavior: SnackBarBehavior.floating,
              ),
            );
          });
        }
        return child ?? const SizedBox.shrink(); // empty widget
      },
    );
  }
}
