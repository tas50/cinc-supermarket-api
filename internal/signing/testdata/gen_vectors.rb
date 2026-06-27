#!/usr/bin/env ruby
# frozen_string_literal: true

# gen_vectors.rb generates mixlib-authentication v1.1 (SHA-1) reference
# signing vectors for the Go test suite. It is the "gold standard" the Go
# implementation is checked against: for a FIXED RSA key, method, path,
# body, user id, and timestamp, it emits the exact X-Ops-* headers and the
# canonical request string that the real Chef server's verifier expects.
#
# Run from internal/signing/testdata/:
#   ruby gen_vectors.rb > vectors.json
#
# Requires the mixlib-authentication gem (gem install mixlib-authentication).

require "json"
require "openssl"
require "base64"
require "digest/sha1"
require "mixlib/authentication/signedheaderauth"

KEY_PATH = File.expand_path("../../../../testdata/test_key.pem", __FILE__)
KEY = OpenSSL::PKey::RSA.new(File.read(KEY_PATH))

def vector(name, method:, path:, body:, user_id:, timestamp:)
  signer = Mixlib::Authentication::SignedHeaderAuth.signing_object(
    http_method: method,
    path: path,
    body: body,
    timestamp: timestamp,
    user_id: user_id,
    proto_version: "1.1"
  )

  # The full set of signed headers, as the client would send them.
  headers = signer.sign(KEY)

  # The canonicalized request string that actually gets RSA-signed.
  canonical = signer.canonicalize_request

  {
    "name"         => name,
    "method"       => method.to_s.upcase,
    "path"         => path,
    "body"         => Base64.strict_encode64(body),
    "user_id"      => user_id,
    "timestamp"    => timestamp,
    "content_hash" => Base64.strict_encode64(Digest::SHA1.digest(body)),
    "canonical"    => canonical,
    "headers"      => headers.transform_keys(&:to_s),
  }
end

vectors = [
  vector("get_no_body",
         method: :get, path: "/api/v1/cookbooks",
         body: "", user_id: "tester", timestamp: "2024-01-01T00:00:00Z"),
  vector("post_json_body",
         method: :post, path: "/api/v1/cookbooks",
         body: %({"a":1}), user_id: "tester", timestamp: "2024-01-01T00:00:00Z"),
  vector("post_tarball_body",
         method: :post, path: "/api/v1/cookbooks",
         body: "FAKE-TAR-BYTES", user_id: "alice", timestamp: "2026-06-27T12:34:56Z"),
  vector("delete_messy_path",
         method: :delete, path: "//api/v1//cookbooks/apache2/",
         body: "", user_id: "bob", timestamp: "2024-12-31T23:59:59Z"),
]

puts JSON.pretty_generate("vectors" => vectors)
