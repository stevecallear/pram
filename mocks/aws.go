// Code generated by MockGen. DO NOT EDIT.
// Source: aws.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	sns "github.com/aws/aws-sdk-go-v2/service/sns"
	sqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	gomock "github.com/golang/mock/gomock"
)

// MockSNS is a mock of SNS interface.
type MockSNS struct {
	ctrl     *gomock.Controller
	recorder *MockSNSMockRecorder
}

// MockSNSMockRecorder is the mock recorder for MockSNS.
type MockSNSMockRecorder struct {
	mock *MockSNS
}

// NewMockSNS creates a new mock instance.
func NewMockSNS(ctrl *gomock.Controller) *MockSNS {
	mock := &MockSNS{ctrl: ctrl}
	mock.recorder = &MockSNSMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSNS) EXPECT() *MockSNSMockRecorder {
	return m.recorder
}

// CreateTopic mocks base method.
func (m *MockSNS) CreateTopic(ctx context.Context, params *sns.CreateTopicInput, optFns ...func(*sns.Options)) (*sns.CreateTopicOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateTopic", varargs...)
	ret0, _ := ret[0].(*sns.CreateTopicOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateTopic indicates an expected call of CreateTopic.
func (mr *MockSNSMockRecorder) CreateTopic(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTopic", reflect.TypeOf((*MockSNS)(nil).CreateTopic), varargs...)
}

// Publish mocks base method.
func (m *MockSNS) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Publish", varargs...)
	ret0, _ := ret[0].(*sns.PublishOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Publish indicates an expected call of Publish.
func (mr *MockSNSMockRecorder) Publish(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockSNS)(nil).Publish), varargs...)
}

// SetTopicAttributes mocks base method.
func (m *MockSNS) SetTopicAttributes(ctx context.Context, params *sns.SetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.SetTopicAttributesOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SetTopicAttributes", varargs...)
	ret0, _ := ret[0].(*sns.SetTopicAttributesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetTopicAttributes indicates an expected call of SetTopicAttributes.
func (mr *MockSNSMockRecorder) SetTopicAttributes(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTopicAttributes", reflect.TypeOf((*MockSNS)(nil).SetTopicAttributes), varargs...)
}

// Subscribe mocks base method.
func (m *MockSNS) Subscribe(ctx context.Context, params *sns.SubscribeInput, optFns ...func(*sns.Options)) (*sns.SubscribeOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(*sns.SubscribeOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockSNSMockRecorder) Subscribe(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockSNS)(nil).Subscribe), varargs...)
}

// MockSQS is a mock of SQS interface.
type MockSQS struct {
	ctrl     *gomock.Controller
	recorder *MockSQSMockRecorder
}

// MockSQSMockRecorder is the mock recorder for MockSQS.
type MockSQSMockRecorder struct {
	mock *MockSQS
}

// NewMockSQS creates a new mock instance.
func NewMockSQS(ctrl *gomock.Controller) *MockSQS {
	mock := &MockSQS{ctrl: ctrl}
	mock.recorder = &MockSQSMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSQS) EXPECT() *MockSQSMockRecorder {
	return m.recorder
}

// CreateQueue mocks base method.
func (m *MockSQS) CreateQueue(ctx context.Context, params *sqs.CreateQueueInput, optFns ...func(*sqs.Options)) (*sqs.CreateQueueOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateQueue", varargs...)
	ret0, _ := ret[0].(*sqs.CreateQueueOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateQueue indicates an expected call of CreateQueue.
func (mr *MockSQSMockRecorder) CreateQueue(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateQueue", reflect.TypeOf((*MockSQS)(nil).CreateQueue), varargs...)
}

// DeleteMessage mocks base method.
func (m *MockSQS) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteMessage", varargs...)
	ret0, _ := ret[0].(*sqs.DeleteMessageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteMessage indicates an expected call of DeleteMessage.
func (mr *MockSQSMockRecorder) DeleteMessage(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMessage", reflect.TypeOf((*MockSQS)(nil).DeleteMessage), varargs...)
}

// GetQueueAttributes mocks base method.
func (m *MockSQS) GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetQueueAttributes", varargs...)
	ret0, _ := ret[0].(*sqs.GetQueueAttributesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetQueueAttributes indicates an expected call of GetQueueAttributes.
func (mr *MockSQSMockRecorder) GetQueueAttributes(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetQueueAttributes", reflect.TypeOf((*MockSQS)(nil).GetQueueAttributes), varargs...)
}

// ReceiveMessage mocks base method.
func (m *MockSQS) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ReceiveMessage", varargs...)
	ret0, _ := ret[0].(*sqs.ReceiveMessageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReceiveMessage indicates an expected call of ReceiveMessage.
func (mr *MockSQSMockRecorder) ReceiveMessage(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReceiveMessage", reflect.TypeOf((*MockSQS)(nil).ReceiveMessage), varargs...)
}

// SetQueueAttributes mocks base method.
func (m *MockSQS) SetQueueAttributes(ctx context.Context, params *sqs.SetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.SetQueueAttributesOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SetQueueAttributes", varargs...)
	ret0, _ := ret[0].(*sqs.SetQueueAttributesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetQueueAttributes indicates an expected call of SetQueueAttributes.
func (mr *MockSQSMockRecorder) SetQueueAttributes(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetQueueAttributes", reflect.TypeOf((*MockSQS)(nil).SetQueueAttributes), varargs...)
}