package minimux_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/meln5674/minimux"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Simple", func() {
	It("should wrap a plain Handler", func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		})

		req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp1 := httptest.NewRecorder()
		mux.ServeHTTP(resp1, req)

		req, err = http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp2 := httptest.NewRecorder()
		Expect(minimux.Simple(mux).ServeHTTP(context.Background(), resp2, req, nil, nil)).ToNot(HaveOccurred())

		Expect(resp1).To(Equal(resp2))
	})
})

var _ = Describe("SimpleFunc", func() {
	It("should wrap a plain HandlerFunc", func() {
		f := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		})

		req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp1 := httptest.NewRecorder()
		f.ServeHTTP(resp1, req)

		req, err = http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp2 := httptest.NewRecorder()
		Expect(minimux.SimpleFunc(f).ServeHTTP(context.Background(), resp2, req, nil, nil)).ToNot(HaveOccurred())

		Expect(resp1).To(Equal(resp2))
	})
})

var _ = Describe("NotFound", func() {
	It("should return 404 and no body", func() {
		req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp := httptest.NewRecorder()
		Expect(minimux.NotFound.ServeHTTP(context.Background(), resp, req, nil, nil)).To(Succeed())
		Expect(resp.Code).To(Equal(http.StatusNotFound))
		Expect(resp.Body.String()).To(BeEmpty())
	})
})

var _ = Describe("Redirecting", func() {
	It("should perform the redirect", func() {
		req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp := httptest.NewRecorder()
		Expect(minimux.RedirectingTo("/foo", http.StatusFound).ServeHTTP(context.Background(), resp, req, nil, nil)).To(Succeed())
		Expect(resp.Code).To(Equal(http.StatusFound))
		location, err := resp.Result().Location()
		Expect(err).ToNot(HaveOccurred())
		Expect(location.String()).To(Equal("/foo"))
	})
})

var _ = Describe("StaticData", func() {
	When("no path variable is specified", func() {
		When("there is data that matches the whole URL", func() {
			It("should return that data", func() {
				s := minimux.StaticData{
					StaticBytes:    map[string]minimux.StaticBytes{"/foo": {Data: []byte("bar"), ContentType: "baz"}},
					DefaultHandler: minimux.NotFound,
				}
				req, err := http.NewRequest(http.MethodGet, "http://localhost/foo", nil)
				Expect(err).ToNot(HaveOccurred())
				resp := httptest.NewRecorder()
				Expect(s.ServeHTTP(context.Background(), resp, req, nil, nil)).To(Succeed())
				res := resp.Result()
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(res.Header).To(HaveKeyWithValue("Content-Type", []string{"baz"}))
				Expect(resp.Body.String()).To(Equal("bar"))
			})
		})
	})
	When("a path variable is specified", func() {
		When("there is data that matches the path variable", func() {
			It("should return that data", func() {
				s := minimux.StaticData{
					StaticBytes:    map[string]minimux.StaticBytes{"/foo": {Data: []byte("bar"), ContentType: "baz"}},
					DefaultHandler: minimux.NotFound,
					PathVar:        "filename",
				}
				req, err := http.NewRequest(http.MethodGet, "http://localhost/prefix/foo", nil)
				Expect(err).ToNot(HaveOccurred())
				resp := httptest.NewRecorder()
				Expect(s.ServeHTTP(context.Background(), resp, req, map[string]string{"filename": "/foo"}, nil)).To(Succeed())
				res := resp.Result()
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(res.Header).To(HaveKeyWithValue("Content-Type", []string{"baz"}))
				Expect(resp.Body.String()).To(Equal("bar"))
			})
		})
	})
	When("There is no data that matches", func() {
		When("There is a default handler", func() {
			It("should call it", func() {
				s := minimux.StaticData{
					StaticBytes:    map[string]minimux.StaticBytes{"foo": {Data: []byte("bar"), ContentType: "baz"}},
					DefaultHandler: minimux.NotFound,
				}
				req, err := http.NewRequest(http.MethodGet, "http://localhost/bar", nil)
				Expect(err).ToNot(HaveOccurred())
				resp := httptest.NewRecorder()
				Expect(s.ServeHTTP(context.Background(), resp, req, nil, nil)).To(Succeed())
				res := resp.Result()
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				Expect(res.Header).ToNot(HaveKey("Content-Type"))
				Expect(resp.Body.String()).To(BeEmpty())
			})
		})
		When("There isn't a default handler", func() {
			It("should do nothing", func() {
				s := minimux.StaticData{
					StaticBytes: map[string]minimux.StaticBytes{"foo": {Data: []byte("bar"), ContentType: "baz"}},
				}
				req, err := http.NewRequest(http.MethodGet, "http://localhost/bar", nil)
				Expect(err).ToNot(HaveOccurred())
				resp := httptest.NewRecorder()
				resp.Header().Add("foo", "bar")
				Expect(s.ServeHTTP(context.Background(), resp, req, nil, nil)).To(Succeed())
				res := resp.Result()
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(res.Header).ToNot(HaveKey("Content-Type"))
				Expect(resp.Body.String()).To(BeEmpty())
			})
		})
	})
})
