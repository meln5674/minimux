package minimux_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/meln5674/minimux"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func stringReader(s string) io.Reader {
	return bytes.NewBuffer([]byte(s))
}

func readString(r io.Reader) (string, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	return buf.String(), err
}

func expectRequest(req *http.Request, method string, path string) {
	GinkgoHelper()
	Expect(req.Method).To(Equal(method), "Method didn't match")
	Expect(req.URL.Path).To(Equal(path), "Path didn't match")
}

func expectRequestWithBody(req *http.Request, method string, path string, body string) {
	GinkgoHelper()
	expectRequest(req, method, path)
	actualBody, err := readString(req.Body)
	Expect(err).ToNot(HaveOccurred(), "Reading request body failed")
	Expect(actualBody).To(Equal(body), "Body didn't match")
}

func expectResponse(handler http.Handler, req *http.Request, statusCode int, body string) {
	GinkgoHelper()
	srv := httptest.NewServer(handler)
	srvURL, err := url.Parse(srv.URL)
	Expect(err).ToNot(HaveOccurred())
	req.URL.Scheme = srvURL.Scheme
	req.URL.Opaque = srvURL.Opaque
	req.URL.User = srvURL.User
	req.URL.Host = srvURL.Host
	req.URL.Path = srvURL.Path + req.URL.Path
	resp, err := srv.Client().Do(req)
	Expect(err).ToNot(HaveOccurred(), "Request failed")
	defer resp.Body.Close()
	Expect(resp.StatusCode).To(Equal(statusCode), "Unexpected status code")
	actualBody, err := readString(resp.Body)
	Expect(err).ToNot(HaveOccurred(), "Reading response body failed")
	Expect(actualBody).To(Equal(body), "Unexpected body")
}

var _ = Describe("A mux", func() {
	DescribeTable(
		"that is empty should return 200 and an empty body for any method or path",
		func(method, path string, body string) {
			req, err := http.NewRequest(method, "http://localhost"+path, stringReader(body))
			Expect(err).ToNot(HaveOccurred())
			expectResponse(&minimux.Mux{}, req, http.StatusOK, "")
		},
		Entry("", http.MethodHead, "/", ""),
		Entry("", http.MethodGet, "/", ""),
		Entry("", http.MethodPost, "/", ""),
		Entry("", http.MethodPut, "/", ""),
		Entry("", http.MethodDelete, "/", ""),
		Entry("", http.MethodHead, "/foo", ""),
		Entry("", http.MethodGet, "/foo", ""),
		Entry("", http.MethodPost, "/foo", ""),
		Entry("", http.MethodPut, "/foo", ""),
		Entry("", http.MethodDelete, "/foo", ""),
		Entry("", http.MethodHead, "/", "body"),
		Entry("", http.MethodGet, "/", "body"),
		Entry("", http.MethodPost, "/", "body"),
		Entry("", http.MethodPut, "/", "body"),
		Entry("", http.MethodDelete, "/", "body"),
	)
	DescribeTable(
		"default handler and no route should call it for any method or path",
		func(method, path string, body string) {
			req, err := http.NewRequest(method, "http://localhost"+path, bytes.NewBuffer([]byte(body)))
			Expect(err).ToNot(HaveOccurred())
			called := false
			expectResponse(&minimux.Mux{
				DefaultHandler: minimux.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
					defer GinkgoRecover()
					called = true
					expectRequestWithBody(req, method, path, body)
					w.WriteHeader(http.StatusOK)
					return nil
				}),
			}, req, http.StatusOK, "")
			Expect(called).To(BeTrue(), "Default handler wasn't called")
		},
		Entry("", http.MethodHead, "/", ""),
		Entry("", http.MethodGet, "/", ""),
		Entry("", http.MethodPost, "/", ""),
		Entry("", http.MethodPut, "/", ""),
		Entry("", http.MethodDelete, "/", ""),
		Entry("", http.MethodHead, "/foo", ""),
		Entry("", http.MethodGet, "/foo", ""),
		Entry("", http.MethodPost, "/foo", ""),
		Entry("", http.MethodPut, "/foo", ""),
		Entry("", http.MethodDelete, "/foo", ""),
		Entry("", http.MethodHead, "/", "body"),
		Entry("", http.MethodGet, "/", "body"),
		Entry("", http.MethodPost, "/", "body"),
		Entry("", http.MethodPut, "/", "body"),
		Entry("", http.MethodDelete, "/", "body"),
	)
	Describe("with a single route", func() {
		var routeCalled bool
		var defaultCalled bool
		var mux *minimux.Mux
		BeforeEach(func() {
			mux = &minimux.Mux{
				Routes: []minimux.Route{
					minimux.
						LiteralPath("/foo").
						WithMethods(http.MethodPost).
						IsHandledBy(minimux.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
							defer GinkgoRecover()
							routeCalled = true
							expectRequestWithBody(req, http.MethodPost, "/foo", "body")
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("resp"))
							return nil
						})),
				},
			}
		})
		Describe("and no default", func() {
			BeforeEach(func() { routeCalled = false })
			It("should call it if it matches", func() {
				req, err := http.NewRequest(http.MethodPost, "http://localhost/foo", stringReader("body"))
				Expect(err).ToNot(HaveOccurred())
				expectResponse(mux, req, http.StatusOK, "resp")
				Expect(routeCalled).To(BeTrue(), "Matching route wasn't called")
			})
			It("should return method not allowed if the path and host match but not the method", func() {
				req, err := http.NewRequest(http.MethodPut, "http://localhost/foo", stringReader("body"))
				Expect(err).ToNot(HaveOccurred())
				expectResponse(mux, req, http.StatusMethodNotAllowed, "")
				Expect(routeCalled).To(BeFalse(), "Matching route was called")
			})
			It("should return OK if the route isn't matched", func() {
				req, err := http.NewRequest(http.MethodPut, "http://localhost/bar", stringReader("body"))
				Expect(err).ToNot(HaveOccurred())
				expectResponse(mux, req, http.StatusOK, "")
				Expect(routeCalled).To(BeFalse(), "Matching route was called")
			})
		})
		Describe("and a default", func() {
			BeforeEach(func() { routeCalled = false; defaultCalled = false })
			BeforeEach(func() {
				mux.DefaultHandler = minimux.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
					defer GinkgoRecover()
					defaultCalled = true
					expectRequestWithBody(req, http.MethodPut, "/bar", "body")
					w.WriteHeader(http.StatusOK)
					return nil
				})
			})
			It("should call it if it matches", func() {
				req, err := http.NewRequest(http.MethodPost, "http://localhost/foo", stringReader("body"))
				Expect(err).ToNot(HaveOccurred())
				expectResponse(mux, req, http.StatusOK, "resp")
				Expect(routeCalled).To(BeTrue(), "Matching route wasn't called")
				Expect(defaultCalled).To(BeFalse(), "Default route was called")
			})
			It("should return method not allowed if the path and host match but not the method", func() {
				req, err := http.NewRequest(http.MethodPut, "http://localhost/foo", stringReader("body"))
				Expect(err).ToNot(HaveOccurred())
				expectResponse(mux, req, http.StatusMethodNotAllowed, "")
				Expect(routeCalled).To(BeFalse(), "Matching route was called")
				Expect(defaultCalled).To(BeFalse(), "Default route was called")
			})
			It("should call the default route isn't matched", func() {
				req, err := http.NewRequest(http.MethodPut, "http://localhost/bar", stringReader("body"))
				Expect(err).ToNot(HaveOccurred())
				expectResponse(mux, req, http.StatusOK, "")
				Expect(defaultCalled).To(BeTrue(), "Default route wasn't called")
				Expect(routeCalled).To(BeFalse(), "Matching route was called")
			})
		})
		Describe("and a pre- and post-processor", func() {
			var preProcessorCalled, routeCalled, postProcessorCalled, deferredFunctionCalled bool
			BeforeEach(func() {
				preProcessorCalled = false
				routeCalled = false
				postProcessorCalled = false
				deferredFunctionCalled = false
			})

			It("should call the pre-processor, the route, the post-processor, then the deferred function, in that order", func() {

				req, err := http.NewRequest(http.MethodGet, "http://localhost/foo", stringReader("body"))
				Expect(err).ToNot(HaveOccurred())
				expectResponse(&minimux.Mux{
					PreProcess: minimux.PreProcessor(func(ctx context.Context, req *http.Request) (context.Context, func()) {
						defer GinkgoRecover()
						preProcessorCalled = true
						expectRequest(req, http.MethodGet, "/foo")
						Expect(routeCalled).To(BeFalse(), "Route was called before PreProcessor")
						Expect(postProcessorCalled).To(BeFalse(), "PostProcessor was called before PostProcessor")
						return ctx, func() {
							defer GinkgoRecover()
							deferredFunctionCalled = true
							Expect(routeCalled).To(BeTrue(), "Defered function was called before Route")
							Expect(postProcessorCalled).To(BeTrue(), "Defered function was called before PostProcessor")
						}
					}),
					PostProcess: minimux.PostProcessor(func(ctx context.Context, req *http.Request, statusCode int, err error) {
						defer GinkgoRecover()
						postProcessorCalled = true
						expectRequest(req, http.MethodGet, "/foo")
						Expect(statusCode).To(Equal(http.StatusNotFound), "Status code was passed to the PostProcessor")
						Expect(err).ToNot(HaveOccurred(), "Unexpected error was passed to PostProcessor")
						Expect(preProcessorCalled).To(BeTrue(), "PostProcessor was called before PreProcessor")
						Expect(routeCalled).To(BeTrue(), "PostProcessor was called before Route")
						Expect(deferredFunctionCalled).To(BeFalse(), "Deferred function was called before PostProcessor")

					}),
					Routes: []minimux.Route{
						minimux.
							LiteralPath("/foo").
							WithMethods(http.MethodGet).
							IsHandledBy(minimux.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
								defer GinkgoRecover()
								routeCalled = true
								expectRequestWithBody(req, http.MethodGet, "/foo", "body")
								w.WriteHeader(http.StatusNotFound)
								w.Write([]byte("resp"))
								return nil
							})),
					},
				}, req, http.StatusNotFound, "resp")
				Expect(preProcessorCalled).To(BeTrue(), "PreProcessor was not called")
				Expect(routeCalled).To(BeTrue(), "Route was not called")
				Expect(postProcessorCalled).To(BeTrue(), "PostProcessor was not called")
				Expect(deferredFunctionCalled).To(BeTrue(), "Deferred function was not called")
			})
			It("should call the post-processor if the pre-processor panics", func() {
				toPanic := fmt.Errorf("This is an error")

				req, err := http.NewRequest(http.MethodGet, "http://localhost/foo", stringReader("body"))
				Expect(err).ToNot(HaveOccurred())
				expectResponse(&minimux.Mux{
					PreProcess: minimux.PreProcessor(func(ctx context.Context, req *http.Request) (context.Context, func()) {
						preProcessorCalled = true
						expectRequest(req, http.MethodGet, "/foo")
						Expect(routeCalled).To(BeFalse(), "Route was called before PreProcessor")
						Expect(postProcessorCalled).To(BeFalse(), "PostProcessor was called before PostProcessor")
						panic(toPanic)
					}),
					PostProcess: minimux.PostProcessor(func(ctx context.Context, req *http.Request, statusCode int, err error) {
						defer GinkgoRecover()
						postProcessorCalled = true
						expectRequest(req, http.MethodGet, "/foo")
						Expect(statusCode).To(Equal(minimux.StatusPreProcessPanic), "Status code was not set to indicate panic")
						Expect(err).To(MatchError(toPanic), "Unexpected error was passed to PostProcessor")
						Expect(preProcessorCalled).To(BeTrue(), "PostProcessor was called before PreProcessor")
						Expect(routeCalled).To(BeFalse(), "Route was called even though PreProcessor panicked")
						Expect(deferredFunctionCalled).To(BeFalse(), "Deferred function was called before PostProcessor")

					}),
					Routes: []minimux.Route{
						minimux.
							LiteralPath("/foo").
							WithMethods(http.MethodGet).
							IsHandledBy(minimux.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
								defer GinkgoRecover()
								routeCalled = true
								expectRequestWithBody(req, http.MethodGet, "/foo", "body")
								w.WriteHeader(http.StatusNotFound)
								w.Write([]byte("resp"))
								return nil
							})),
					},
				}, req, http.StatusInternalServerError, "")
				Expect(preProcessorCalled).To(BeTrue(), "PreProcessor was not called")
				Expect(routeCalled).To(BeFalse(), "Route was called")
				Expect(postProcessorCalled).To(BeTrue(), "PostProcessor was not called")
				Expect(deferredFunctionCalled).To(BeFalse(), "Deferred function was called")
			})
		})
	})
	Describe("with a post-processor", func() {
		It("should call the post-processor if the route panics", func() {
			routeCalled := false
			postProcessorCalled := false
			req, err := http.NewRequest(http.MethodGet, "http://localhost/foo", stringReader("body"))
			Expect(err).ToNot(HaveOccurred())
			expectResponse(&minimux.Mux{
				PostProcess: minimux.PostProcessor(func(ctx context.Context, req *http.Request, statusCode int, err error) {
					defer GinkgoRecover()
					postProcessorCalled = true
					expectRequest(req, http.MethodGet, "/foo")
					Expect(statusCode).To(Equal(minimux.StatusPanic), "Status code was not set to indicate a panic")
					Expect(err).To(HaveOccurred(), "Panicked value was not passed to PostProcessor")
					Expect(routeCalled).To(BeTrue(), "PostProcessor was called before Route")

				}),
				Routes: []minimux.Route{
					minimux.
						LiteralPath("/foo").
						WithMethods(http.MethodGet).
						IsHandledBy(minimux.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
							routeCalled = true
							expectRequestWithBody(req, http.MethodGet, "/foo", "body")
							w.WriteHeader(http.StatusNotFound)
							w.Write([]byte("resp"))
							// Deliberate index-out-of-bounds to trigger panic
							_ = w.Header()["foo"][10]
							return nil
						})),
				},
			}, req, http.StatusNotFound, "resp")
			Expect(routeCalled).To(BeTrue(), "Route was not called")
			Expect(postProcessorCalled).To(BeTrue(), "PostProcessor was not called")
		})
	})
	Describe("nested in another mux with a prefix", func() {
		It("should pass down any path variables and strip the prefix", func() {
			routeCalled := false
			req, err := http.NewRequest(http.MethodGet, "http://localhost/foo/bar", stringReader("body"))
			Expect(err).ToNot(HaveOccurred())
			expectResponse(&minimux.Mux{
				Routes: []minimux.Route{
					minimux.
						PathWithVars("/foo(/.*)", "suffix").
						IsHandledBy(minimux.InnerMuxWithPrefix("suffix", &minimux.Mux{
							Routes: []minimux.Route{
								minimux.
									LiteralPath("/bar").
									WithMethods(http.MethodGet).
									IsHandledBy(minimux.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
										defer GinkgoRecover()
										routeCalled = true
										expectRequestWithBody(req, http.MethodGet, "/bar", "body")
										// Deliberatly not writing the header here
										w.Write([]byte("resp"))
										return nil
									})),
							},
						})),
				},
			}, req, http.StatusOK, "resp")
			Expect(routeCalled).To(BeTrue(), "Route was not called")
		})
	})
	Describe("nested in another mux without a prefix", func() {
		It("should pass down any path variables and keep the prefix", func() {
			routeCalled := false
			req, err := http.NewRequest(http.MethodGet, "http://localhost/foo/bar", stringReader("body"))
			Expect(err).ToNot(HaveOccurred())
			expectResponse(&minimux.Mux{
				Routes: []minimux.Route{
					minimux.
						PathPattern("/foo/.*").
						IsHandledBy(minimux.InnerMux(&minimux.Mux{
							Routes: []minimux.Route{
								minimux.
									LiteralPath("/foo/bar").
									WithMethods(http.MethodGet).
									WithForm().
									IsHandledByFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
										defer GinkgoRecover()
										routeCalled = true
										expectRequestWithBody(req, http.MethodGet, "/foo/bar", "body")
										// Deliberatly not writing the header here
										w.Write([]byte("resp"))
										return nil
									}),
							},
						})),
				},
			}, req, http.StatusOK, "resp")
			Expect(routeCalled).To(BeTrue(), "Route was not called")
		})
	})
	Describe("with a route that has a form", func() {
		It("Should parse the form", func() {
			routeCalled := false
			req, err := http.NewRequest(http.MethodGet, "http://localhost/foo?bar=qux", stringReader("body"))
			Expect(err).ToNot(HaveOccurred())
			expectResponse(&minimux.Mux{
				Routes: []minimux.Route{
					minimux.
						LiteralPath("/foo").
						WithForm().
						IsHandledBy(minimux.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, pathVars map[string]string, formErr error) error {
							defer GinkgoRecover()
							routeCalled = true
							expectRequestWithBody(req, http.MethodGet, "/foo", "body")
							Expect(req.Form.Get("bar")).To(Equal("qux"), "Form was not parsed")
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("resp"))
							return nil
						})),
				},
			}, req, http.StatusOK, "resp")
			Expect(routeCalled).To(BeTrue(), "Route was not called")
		})
	})
})
