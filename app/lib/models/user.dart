import 'package:json_annotation/json_annotation.dart';

@JsonSerializable()
class User {
  final int id;
  // final ByteData userSecret;

  User({required this.id});

  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
}
