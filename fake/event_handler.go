// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"context"
	"sync"

	"github.com/phogolabs/theta"
)

type EventHandler struct {
	HandleContextStub        func(context.Context, *theta.EventArgs) error
	handleContextMutex       sync.RWMutex
	handleContextArgsForCall []struct {
		arg1 context.Context
		arg2 *theta.EventArgs
	}
	handleContextReturns struct {
		result1 error
	}
	handleContextReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *EventHandler) HandleContext(arg1 context.Context, arg2 *theta.EventArgs) error {
	fake.handleContextMutex.Lock()
	ret, specificReturn := fake.handleContextReturnsOnCall[len(fake.handleContextArgsForCall)]
	fake.handleContextArgsForCall = append(fake.handleContextArgsForCall, struct {
		arg1 context.Context
		arg2 *theta.EventArgs
	}{arg1, arg2})
	fake.recordInvocation("HandleContext", []interface{}{arg1, arg2})
	fake.handleContextMutex.Unlock()
	if fake.HandleContextStub != nil {
		return fake.HandleContextStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.handleContextReturns
	return fakeReturns.result1
}

func (fake *EventHandler) HandleContextCallCount() int {
	fake.handleContextMutex.RLock()
	defer fake.handleContextMutex.RUnlock()
	return len(fake.handleContextArgsForCall)
}

func (fake *EventHandler) HandleContextCalls(stub func(context.Context, *theta.EventArgs) error) {
	fake.handleContextMutex.Lock()
	defer fake.handleContextMutex.Unlock()
	fake.HandleContextStub = stub
}

func (fake *EventHandler) HandleContextArgsForCall(i int) (context.Context, *theta.EventArgs) {
	fake.handleContextMutex.RLock()
	defer fake.handleContextMutex.RUnlock()
	argsForCall := fake.handleContextArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *EventHandler) HandleContextReturns(result1 error) {
	fake.handleContextMutex.Lock()
	defer fake.handleContextMutex.Unlock()
	fake.HandleContextStub = nil
	fake.handleContextReturns = struct {
		result1 error
	}{result1}
}

func (fake *EventHandler) HandleContextReturnsOnCall(i int, result1 error) {
	fake.handleContextMutex.Lock()
	defer fake.handleContextMutex.Unlock()
	fake.HandleContextStub = nil
	if fake.handleContextReturnsOnCall == nil {
		fake.handleContextReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.handleContextReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *EventHandler) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.handleContextMutex.RLock()
	defer fake.handleContextMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *EventHandler) recordInvocation(key string, args []interface{}) {
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

var _ theta.EventHandler = new(EventHandler)
