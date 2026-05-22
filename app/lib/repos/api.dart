import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;
import 'package:logging/logging.dart';

class Api {
  final String baseUrl;
  final Logger _logger = Logger("API");

  Api({required this.baseUrl});

  static Api main() {
    final baseUrl = const String.fromEnvironment(
      'API_URL',
      // defaultValue: "https://api.thecoven.space",
    );
    assert(baseUrl != "");
    return Api(baseUrl: baseUrl);
  }

  Future<ApiResponse> request(
    String method,
    String path, {
    Map<String, dynamic>? queryParams,
    Map<String, String> headers = const {},
    Object? body,
  }) async {
    final url = _uri(path, queryParams: queryParams);
    final req = http.Request(method, url);
    req.headers.addAll(headers);
    if (body is String) {
      req.body = body;
    } else {
      // TODO: other types?
      req.body = jsonEncode(body);
      req.headers['Content-Type'] = "application/json";
    }
    final res = await req.send();
    _logger.log(
      Level.INFO,
      "API: ${method.toUpperCase()} $url -- [${res.statusCode}] size:${res.contentLength}",
    );
    final resBody = await res.stream.bytesToString();
    return ApiResponse(res, resBody);
  }

  Uri _uri(String path, {Map<String, dynamic>? queryParams}) {
    return Uri.parse(baseUrl + path).replace(queryParameters: queryParams);
  }
}

class ApiResponse {
  final http.StreamedResponse _response;
  final String _responseBody;

  ApiResponse(this._response, this._responseBody);

  ApiSingleResponse<T> single<T>({
    required T Function(Map<String, dynamic>) fromJson,
  }) {
    if (_response.statusCode >= 400) {
      throw _error();
    }
    final body = _parseBody();
    assert(body["status"] is String);
    assert(body["data"] is Map<String, dynamic>);
    return ApiSingleResponse(body["status"], fromJson(body["data"]));
  }

  ApiPageResponse<T> many<T>({
    required T Function(Map<String, dynamic>) fromJson,
  }) {
    if (_response.statusCode >= 400) {
      throw _error();
    }
    final body = _parseBody();
    Pagination? pagination;
    if (body["pagination"] != null) {
      assert(body["pagination"] is Map<String, dynamic>);
      pagination = Pagination.fromJson(body["pagination"]);
    }
    assert(body["status"] is String);
    assert(body["data"] is List<Map<String, dynamic>>);
    List<T> data = body["data"].map((v) => fromJson(v)).toList();
    return ApiPageResponse(body["status"], data, pagination);
  }

  ApiException _error() {
    if (_response.headers["content-type"] != "application/json") {
      return ApiException(
        "HTTP status ${_response.statusCode}",
        response: _response,
      );
    }
    final body = _parseBody();
    assert(body["status"] == "ERROR");
    assert(body["error"] is Map<String, dynamic>);
    String message = body["error"]["message"];
    return ApiException(message, response: _response, body: body);
  }

  Map<String, dynamic> _parseBody() {
    return json.decode(_responseBody) as Map<String, dynamic>;
  }
}

class ApiSingleResponse<T> {
  final String status;
  final T data;
  ApiSingleResponse(this.status, this.data);
}

class ApiPageResponse<T> {
  final String status;
  final List<T> data;
  final Pagination? pagination;

  ApiPageResponse(this.status, this.data, this.pagination);
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
  final http.StreamedResponse response;
  final Map<String, dynamic>? body;

  ApiException(super.message, {required this.response, this.body})
    : super(uri: response.request?.url);
}
