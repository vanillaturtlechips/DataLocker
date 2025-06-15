// Package model provides database models for DataLocker application.
// This file defines custom errors for model validation and operations.
package model

import "errors"

// File 모델 관련 에러
var (
	// ErrEmptyOriginalName 원본 파일명이 비어있음
	ErrEmptyOriginalName = errors.New("원본 파일명은 필수입니다")

	// ErrOriginalNameTooLong 원본 파일명이 너무 김
	ErrOriginalNameTooLong = errors.New("원본 파일명이 너무 깁니다")

	// ErrEmptyEncryptedPath 암호화된 파일 경로가 비어있음
	ErrEmptyEncryptedPath = errors.New("암호화된 파일 경로는 필수입니다")

	// ErrEncryptedPathTooLong 암호화된 파일 경로가 너무 김
	ErrEncryptedPathTooLong = errors.New("암호화된 파일 경로가 너무 깁니다")

	// ErrInvalidFileSize 잘못된 파일 크기
	ErrInvalidFileSize = errors.New("파일 크기는 0 이상이어야 합니다")

	// ErrEmptyMimeType MIME 타입이 비어있음
	ErrEmptyMimeType = errors.New("MIME 타입은 필수입니다")

	// ErrMimeTypeTooLong MIME 타입이 너무 김
	ErrMimeTypeTooLong = errors.New("MIME 타입이 너무 깁니다")

	// ErrEmptyChecksum 체크섬이 비어있음
	ErrEmptyChecksum = errors.New("체크섬은 필수입니다")

	// ErrChecksumTooLong 체크섬이 너무 김
	ErrChecksumTooLong = errors.New("체크섬이 너무 깁니다")

	// ErrInvalidFileStatus 잘못된 파일 상태
	ErrInvalidFileStatus = errors.New("잘못된 파일 상태입니다")
)

// EncryptionMetadata 모델 관련 에러
var (
	// ErrInvalidFileID 잘못된 파일 ID
	ErrInvalidFileID = errors.New("유효하지 않은 파일 ID입니다")

	// ErrEmptyAlgorithm 암호화 알고리즘이 비어있음
	ErrEmptyAlgorithm = errors.New("암호화 알고리즘은 필수입니다")

	// ErrAlgorithmTooLong 암호화 알고리즘이 너무 김
	ErrAlgorithmTooLong = errors.New("암호화 알고리즘명이 너무 깁니다")

	// ErrInvalidAlgorithm 지원하지 않는 암호화 알고리즘
	ErrInvalidAlgorithm = errors.New("지원하지 않는 암호화 알고리즘입니다")

	// ErrEmptyKeyDerivation 키 유도 방식이 비어있음
	ErrEmptyKeyDerivation = errors.New("키 유도 방식은 필수입니다")

	// ErrKeyDerivationTooLong 키 유도 방식이 너무 김
	ErrKeyDerivationTooLong = errors.New("키 유도 방식명이 너무 깁니다")

	// ErrInvalidKeyDerivation 지원하지 않는 키 유도 방식
	ErrInvalidKeyDerivation = errors.New("지원하지 않는 키 유도 방식입니다")

	// ErrEmptySalt Salt가 비어있음
	ErrEmptySalt = errors.New("salt는 필수입니다")

	// ErrSaltTooLong Salt가 너무 김
	ErrSaltTooLong = errors.New("salt가 너무 깁니다")

	// ErrInvalidSaltHex 잘못된 Salt hex 형식
	ErrInvalidSaltHex = errors.New("잘못된 salt hex 형식입니다")

	// ErrInvalidSaltSize 잘못된 Salt 크기
	ErrInvalidSaltSize = errors.New("salt 크기가 올바르지 않습니다")

	// ErrEmptyNonce Nonce가 비어있음
	ErrEmptyNonce = errors.New("nonce는 필수입니다")

	// ErrNonceTooLong Nonce가 너무 김
	ErrNonceTooLong = errors.New("nonce가 너무 깁니다")

	// ErrInvalidNonceHex 잘못된 Nonce hex 형식
	ErrInvalidNonceHex = errors.New("잘못된 nonce hex 형식입니다")

	// ErrInvalidNonceSize 잘못된 Nonce 크기
	ErrInvalidNonceSize = errors.New("nonce 크기가 올바르지 않습니다")

	// ErrInvalidIterations 잘못된 반복 횟수
	ErrInvalidIterations = errors.New("반복 횟수는 1,000 이상 1,000,000 이하여야 합니다")
)

// 일반적인 모델 에러
var (
	// ErrRecordNotFound 레코드를 찾을 수 없음
	ErrRecordNotFound = errors.New("레코드를 찾을 수 없습니다")

	// ErrDuplicateRecord 중복된 레코드
	ErrDuplicateRecord = errors.New("중복된 레코드입니다")

	// ErrForeignKeyConstraint 외래키 제약조건 위반
	ErrForeignKeyConstraint = errors.New("외래키 제약조건 위반입니다")

	// ErrInvalidModelData 잘못된 모델 데이터
	ErrInvalidModelData = errors.New("잘못된 모델 데이터입니다")
)
