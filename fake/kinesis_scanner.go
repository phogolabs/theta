// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"context"
	"sync"

	consumer "github.com/harlow/kinesis-consumer"
	"github.com/phogolabs/theta"
)

type KinesisScanner struct {
	ScanStub        func(context.Context, consumer.ScanFunc) error
	scanMutex       sync.RWMutex
	scanArgsForCall []struct {
		arg1 context.Context
		arg2 consumer.ScanFunc
	}
	scanReturns struct {
		result1 error
	}
	scanReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *KinesisScanner) Scan(arg1 context.Context, arg2 consumer.ScanFunc) error {
	fake.scanMutex.Lock()
	ret, specificReturn := fake.scanReturnsOnCall[len(fake.scanArgsForCall)]
	fake.scanArgsForCall = append(fake.scanArgsForCall, struct {
		arg1 context.Context
		arg2 consumer.ScanFunc
	}{arg1, arg2})
	fake.recordInvocation("Scan", []interface{}{arg1, arg2})
	fake.scanMutex.Unlock()
	if fake.ScanStub != nil {
		return fake.ScanStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.scanReturns
	return fakeReturns.result1
}

func (fake *KinesisScanner) ScanCallCount() int {
	fake.scanMutex.RLock()
	defer fake.scanMutex.RUnlock()
	return len(fake.scanArgsForCall)
}

func (fake *KinesisScanner) ScanCalls(stub func(context.Context, consumer.ScanFunc) error) {
	fake.scanMutex.Lock()
	defer fake.scanMutex.Unlock()
	fake.ScanStub = stub
}

func (fake *KinesisScanner) ScanArgsForCall(i int) (context.Context, consumer.ScanFunc) {
	fake.scanMutex.RLock()
	defer fake.scanMutex.RUnlock()
	argsForCall := fake.scanArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *KinesisScanner) ScanReturns(result1 error) {
	fake.scanMutex.Lock()
	defer fake.scanMutex.Unlock()
	fake.ScanStub = nil
	fake.scanReturns = struct {
		result1 error
	}{result1}
}

func (fake *KinesisScanner) ScanReturnsOnCall(i int, result1 error) {
	fake.scanMutex.Lock()
	defer fake.scanMutex.Unlock()
	fake.ScanStub = nil
	if fake.scanReturnsOnCall == nil {
		fake.scanReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.scanReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *KinesisScanner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.scanMutex.RLock()
	defer fake.scanMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *KinesisScanner) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ theta.KinesisScanner = new(KinesisScanner)