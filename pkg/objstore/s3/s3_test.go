// Copyright (c) The Thanos Authors.
// Licensed under the Apache License 2.0.

package s3

import (
	"testing"
	"time"

	"github.com/thanos-io/thanos/pkg/testutil"
)

func TestParseConfig(t *testing.T) {
	input := []byte(`bucket: abcd
insecure: false`)
	cfg, err := parseConfig(input)
	testutil.Ok(t, err)

	if cfg.Bucket != "abcd" {
		t.Errorf("parsing of bucket failed: got %v, expected %v", cfg.Bucket, "abcd")
	}
	if cfg.Insecure {
		t.Errorf("parsing of insecure failed: got %v, expected %v", cfg.Insecure, false)
	}
}

func TestParseConfig_SSEConfig(t *testing.T) {
	input := []byte(`bucket: abdd
endpoint: "s3-endpoint"
sse_config:
  type: SSE-S3`)

	cfg, err := parseConfig(input)
	testutil.Ok(t, err)
	testutil.Ok(t, validate(cfg))

	input2 := []byte(`bucket: abdd
endpoint: "s3-endpoint"
sse_config:
  type: SSE-C`)

	cfg, err = parseConfig(input2)
	testutil.Ok(t, err)
	testutil.NotOk(t, validate(cfg))

	input3 := []byte(`bucket: abdd
endpoint: "s3-endpoint"
sse_config:
  type: SSE-C
  kms_key_id: qweasd`)

	cfg, err = parseConfig(input3)
	testutil.Ok(t, err)
	testutil.NotOk(t, validate(cfg))

	input4 := []byte(`bucket: abdd
endpoint: "s3-endpoint"
sse_config:
  type: SSE-C
  encryption_key: /some/file`)

	cfg, err = parseConfig(input4)
	testutil.Ok(t, err)
	testutil.Ok(t, validate(cfg))

	input5 := []byte(`bucket: abdd
endpoint: "s3-endpoint"
sse_config:
  type: SSE-KMS`)

	cfg, err = parseConfig(input5)
	testutil.Ok(t, err)
	testutil.NotOk(t, validate(cfg))

	input6 := []byte(`bucket: abdd
endpoint: "s3-endpoint"
sse_config:
  type: SSE-KMS
  kms_key_id: abcd1234-ab12-cd34-1234567890ab`)

	cfg, err = parseConfig(input6)
	testutil.Ok(t, err)
	testutil.Ok(t, validate(cfg))

	input7 := []byte(`bucket: abdd
endpoint: "s3-endpoint"
sse_config:
  type: SSE-KMS
  kms_key_id: abcd1234-ab12-cd34-1234567890ab
  kms_encryption_context:
    key: value
    something: else
    a: b`)

	cfg, err = parseConfig(input7)
	testutil.Ok(t, err)
	testutil.Ok(t, validate(cfg))

	input8 := []byte(`bucket: abdd
endpoint: "s3-endpoint"
sse_config:
  type: SSE-MagicKey
  kms_key_id: abcd1234-ab12-cd34-1234567890ab
  encryption_key: /some/file`)

	cfg, err = parseConfig(input8)
	testutil.Ok(t, err)
	// Since the error handling for "proper type" if done as we're setting up the bucket.
	testutil.Ok(t, validate(cfg))
}

func TestParseConfig_DefaultHTTPConfig(t *testing.T) {
	input := []byte(`bucket: abcd
insecure: false`)
	cfg, err := parseConfig(input)
	testutil.Ok(t, err)

	if time.Duration(cfg.HTTPConfig.IdleConnTimeout) != time.Duration(90*time.Second) {
		t.Errorf("parsing of idle_conn_timeout failed: got %v, expected %v",
			time.Duration(cfg.HTTPConfig.IdleConnTimeout), time.Duration(90*time.Second))
	}

	if time.Duration(cfg.HTTPConfig.ResponseHeaderTimeout) != time.Duration(2*time.Minute) {
		t.Errorf("parsing of response_header_timeout failed: got %v, expected %v",
			time.Duration(cfg.HTTPConfig.IdleConnTimeout), time.Duration(2*time.Minute))
	}

	if cfg.HTTPConfig.InsecureSkipVerify {
		t.Errorf("parsing of insecure_skip_verify failed: got %v, expected %v", cfg.HTTPConfig.InsecureSkipVerify, false)
	}
}

func TestParseConfig_CustomHTTPConfig(t *testing.T) {
	input := []byte(`bucket: abcd
insecure: false
http_config:
  insecure_skip_verify: true
  idle_conn_timeout: 50s
  response_header_timeout: 1m`)
	cfg, err := parseConfig(input)
	testutil.Ok(t, err)

	if time.Duration(cfg.HTTPConfig.IdleConnTimeout) != time.Duration(50*time.Second) {
		t.Errorf("parsing of idle_conn_timeout failed: got %v, expected %v",
			time.Duration(cfg.HTTPConfig.IdleConnTimeout), time.Duration(50*time.Second))
	}

	if time.Duration(cfg.HTTPConfig.ResponseHeaderTimeout) != time.Duration(1*time.Minute) {
		t.Errorf("parsing of response_header_timeout failed: got %v, expected %v",
			time.Duration(cfg.HTTPConfig.IdleConnTimeout), time.Duration(1*time.Minute))
	}

	if !cfg.HTTPConfig.InsecureSkipVerify {
		t.Errorf("parsing of insecure_skip_verify failed: got %v, expected %v", cfg.HTTPConfig.InsecureSkipVerify, false)
	}
}

func TestValidate_OK(t *testing.T) {
	input := []byte(`bucket: "bucket-name"
endpoint: "s3-endpoint"
access_key: "access_key"
insecure: false
signature_version2: false
secret_key: "secret_key"
http_config:
  insecure_skip_verify: false
  idle_conn_timeout: 50s`)
	cfg, err := parseConfig(input)
	testutil.Ok(t, err)
	testutil.Ok(t, validate(cfg))
	testutil.Assert(t, cfg.PutUserMetadata != nil, "map should not be nil")

	input2 := []byte(`bucket: "bucket-name"
endpoint: "s3-endpoint"
access_key: "access_key"
insecure: false
signature_version2: false
secret_key: "secret_key"
put_user_metadata:
  "X-Amz-Acl": "bucket-owner-full-control"
http_config:
  idle_conn_timeout: 0s`)
	cfg2, err := parseConfig(input2)
	testutil.Ok(t, err)
	testutil.Ok(t, validate(cfg2))

	testutil.Equals(t, "bucket-owner-full-control", cfg2.PutUserMetadata["X-Amz-Acl"])
}

func TestParseConfig_PartSize(t *testing.T) {
	input := []byte(`bucket: "bucket-name"
endpoint: "s3-endpoint"
access_key: "access_key"
insecure: false
signature_version2: false
secret_key: "secret_key"
http_config:
  insecure_skip_verify: false
  idle_conn_timeout: 50s`)

	cfg, err := parseConfig(input)
	testutil.Ok(t, err)
	testutil.Assert(t, cfg.PartSize == 1024*1024*128, "when part size not set it should default to 128MiB")

	input2 := []byte(`bucket: "bucket-name"
endpoint: "s3-endpoint"
access_key: "access_key"
insecure: false
signature_version2: false
secret_key: "secret_key"
part_size: 104857600
http_config:
  insecure_skip_verify: false
  idle_conn_timeout: 50s`)
	cfg2, err := parseConfig(input2)
	testutil.Ok(t, err)
	testutil.Assert(t, cfg2.PartSize == 1024*1024*100, "when part size should be set to 100MiB")
}

func TestParseConfig_OldSEEncryptionFieldShouldFail(t *testing.T) {
	input := []byte(`bucket: "bucket-name"
endpoint: "s3-endpoint"
access_key: "access_key"
insecure: false
signature_version2: false
encrypt_sse: false
secret_key: "secret_key"
see_encryption: true
put_user_metadata:
  "X-Amz-Acl": "bucket-owner-full-control"
http_config:
  idle_conn_timeout: 0s`)
	_, err := parseConfig(input)
	testutil.NotOk(t, err)
}

func TestParseConfig_ListObjectsV1(t *testing.T) {
	input := []byte(`bucket: "bucket-name"
endpoint: "s3-endpoint"`)

	cfg, err := parseConfig(input)
	testutil.Ok(t, err)

	if cfg.ListObjectsVersion != "" {
		t.Errorf("when list_objects_version not set, it should default to empty")
	}

	input2 := []byte(`bucket: "bucket-name"
endpoint: "s3-endpoint"
list_objects_version: "abcd"`)

	cfg2, err := parseConfig(input2)
	testutil.Ok(t, err)

	if cfg2.ListObjectsVersion != "abcd" {
		t.Errorf("parsing of list_objects_version failed: got %v, expected %v", cfg.ListObjectsVersion, "abcd")
	}
}
