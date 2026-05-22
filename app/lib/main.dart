import 'dart:io';
import 'dart:ui';

import 'package:flutter/foundation.dart';
import 'package:logging/logging.dart';
import 'package:flutter/material.dart';

import './ui/splash.dart';

void main() {
  Logger.root.level = Level.ALL; // defaults to Level.INFO
  Logger.root.onRecord.listen((record) {
    print('${record.level.name}: ${record.time}: ${record.message}');
  });

  if (kDebugMode) {
    // add a global unhandled exception handler that crashes the app, forcing resolution
    PlatformDispatcher.instance.onError = (err, stacktrace) {
      Logger.root.log(Level.SHOUT, "Unhandled exception: $err\n$stacktrace");
      exit(1);
    };
  }

  runApp(const CovenApp());
}

class CovenApp extends StatelessWidget {
  const CovenApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'The Coven',
      theme: ThemeData(
        colorScheme: .fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: SplashPage(),
    );
  }
}
