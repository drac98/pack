// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/buildpacks/pack/pkg/client (interfaces: BuildpackDownloader)

// Package testmocks is a generated GoMock package.
package testmocks

import (
	context "context"
	reflect "reflect"

	buildpack "github.com/buildpacks/pack/pkg/buildpack"
	gomock "github.com/golang/mock/gomock"
)

// MockBuildpackDownloader is a mock of BuildpackDownloader interface.
type MockBuildpackDownloader struct {
	ctrl     *gomock.Controller
	recorder *MockBuildpackDownloaderMockRecorder
}

// MockBuildpackDownloaderMockRecorder is the mock recorder for MockBuildpackDownloader.
type MockBuildpackDownloaderMockRecorder struct {
	mock *MockBuildpackDownloader
}

// NewMockBuildpackDownloader creates a new mock instance.
func NewMockBuildpackDownloader(ctrl *gomock.Controller) *MockBuildpackDownloader {
	mock := &MockBuildpackDownloader{ctrl: ctrl}
	mock.recorder = &MockBuildpackDownloaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBuildpackDownloader) EXPECT() *MockBuildpackDownloaderMockRecorder {
	return m.recorder
}

// Download mocks base method.
func (m *MockBuildpackDownloader) Download(arg0 context.Context, arg1 string, arg2 buildpack.DownloadOptions) (buildpack.BuildModule, []buildpack.BuildModule, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Download", arg0, arg1, arg2)
	ret0, _ := ret[0].(buildpack.BuildModule)
	ret1, _ := ret[1].([]buildpack.BuildModule)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Download indicates an expected call of Download.
func (mr *MockBuildpackDownloaderMockRecorder) Download(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Download", reflect.TypeOf((*MockBuildpackDownloader)(nil).Download), arg0, arg1, arg2)
}
