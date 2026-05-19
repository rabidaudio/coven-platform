import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;

class ApiResponse<T> {
  final String status;
  final T data;

  ApiResponse(this.status, this.data);

  factory ApiResponse.fromResponse(
    http.Response response,
    T Function(Map<String, dynamic>) fromJson,
  ) {
    if (response.statusCode >= 400) {
      throw ApiException.fromResponse(response);
    }
    Map<String, dynamic> body =
        json.decode(response.body) as Map<String, dynamic>;

    return ApiResponse(body["status"], fromJson(body["data"]));
  }
}

class ApiPageResponse<T> {
  final String status;
  final List<T> data;
  final Pagination? pagination;

  ApiPageResponse(this.status, this.data, this.pagination);

  factory ApiPageResponse.fromResponse(
    http.Response response,
    T Function(Map<String, dynamic>) fromJson,
  ) {
    if (response.statusCode >= 400) {
      throw ApiException.fromResponse(response);
    }
    Map<String, dynamic> body =
        json.decode(response.body) as Map<String, dynamic>;
    Pagination? pagination;
    if (body["pagination"] != null) {
      pagination = Pagination.fromJson(body["pagination"]);
    }

    assert(body["data"] is List<Map<String, dynamic>>);
    List<T> data = body["data"].map((v) => fromJson(v)).toList();
    return ApiPageResponse(body["status"], data, pagination);
  }
}

class Pagination {
  final int page;
  final int perPage;
  final int totalItems;

  Pagination(this.page, this.perPage, this.totalItems);

  factory Pagination.fromJson(Map<String, dynamic> pagination) {
    return Pagination(
      pagination["page"],
      pagination["per_page"],
      pagination["total_items"],
    );
  }
}

class ApiException extends HttpException {
  final http.Response response;
  final Map<String, dynamic>? body;

  ApiException(
    String message, {
    required http.Response this.response,
    this.body,
  }) : super(message, uri: response.request?.url);

  factory ApiException.fromResponse(var response) {
    if (response.headers["content-type"] != "application/json") {
      return ApiException(
        "HTTP status ${response.statusCode}",
        response: response,
      );
    }
    Map<String, dynamic> body =
        json.decode(response.body) as Map<String, dynamic>;
    String message = body["error"]["message"];
    return ApiException(message, response: response, body: body);
  }
}

class Api {
  final String baseUrl;

  Api({required this.baseUrl});

  static Api main() {
    final baseUrl = String.fromEnvironment(
      'API_URL',
      defaultValue: "https://api.thecoven.space",
    );
    return Api(baseUrl: baseUrl);
  }

  Future<ApiResponse<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParams,
    required T Function(Map<String, dynamic>) fromJson,
  }) async {
    var res = await http.get(_uri(path, queryParams: queryParams));
    return ApiResponse.fromResponse(res, fromJson);
  }

  Future<ApiPageResponse<T>> getMany<T>(
    String path, {
    Map<String, dynamic>? queryParams,
    required T Function(Map<String, dynamic>) fromJson,
  }) async {
    var res = await http.get(_uri(path, queryParams: queryParams));
    return ApiPageResponse.fromResponse(res, fromJson);
  }

  Future<ApiResponse<T>> put<T>(
    String path, {
    Map<String, dynamic>? queryParams,
    Object? body,
    required T Function(Map<String, dynamic>) fromJson,
  }) async {
    Map<String, String> headers = {};
    var bodyStr = "";
    if (body != null) {
      bodyStr = jsonEncode(body);
      headers['Content-Type'] = "application/json";
    }
    var res = await http.put(
      _uri(path, queryParams: queryParams),
      body: bodyStr,
      headers: headers,
    );
    return ApiResponse.fromResponse(res, fromJson);
  }

  Future<ApiResponse<T>> post<T>(
    String path, {
    Map<String, dynamic>? queryParams,
    Object? body,
    required T Function(Map<String, dynamic>) fromJson,
  }) async {
    Map<String, String> headers = {};
    var bodyStr = "";
    if (body != null) {
      bodyStr = jsonEncode(body);
      headers['Content-Type'] = "application/json";
    }
    var res = await http.post(
      _uri(path, queryParams: queryParams),
      body: bodyStr,
      headers: headers,
    );
    return ApiResponse.fromResponse(res, fromJson);
  }

  Future<ApiResponse<T>> delete<T>(
    String path, {
    Map<String, dynamic>? queryParams,
    Object? body,
    required T Function(Map<String, dynamic>) fromJson,
  }) async {
    var res = await http.delete(
      _uri(path, queryParams: queryParams),
      body: body,
    );
    return ApiResponse.fromResponse(res, fromJson);
  }

  Uri _uri(String path, {Map<String, dynamic>? queryParams}) {
    return Uri.parse(baseUrl + path).replace(queryParameters: queryParams);
  }
}
