package broker

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	meshplayv1alpha1 "github.com/khulnasoft/meshplay-operator/api/v1alpha1"
)

// Initialize test suite entrypoint
func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Broker Suite")
}

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
	clientSet *kubernetes.Clientset
)

var _ = BeforeSuite(func(ctx SpecContext) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	timeout := 3 * time.Minute
	testEnv = &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
		},
		ControlPlaneStartTimeout: timeout,
		ControlPlaneStopTimeout:  timeout,
		AttachControlPlaneOutput: false,
		BinaryAssetsDirectory:    filepath.Join("..", "..", "bin", "k8s", "1.24.2-linux-amd64"),
	}

	var cfg *rest.Config
	var err error
	done := make(chan interface{})
	go func() {
		defer GinkgoRecover()
		cfg, err = testEnv.Start()
		close(done)
	}()
	Eventually(done).WithContext(ctx).WithTimeout(timeout).Should(BeClosed())
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := runtime.NewScheme()

	Expect(meshplayv1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(k8sscheme.AddToScheme(scheme)).To(Succeed())
	Expect(apiv1.AddToScheme(scheme)).To(Succeed())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	clientSet, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(clientSet).NotTo(BeNil())

	crd := &apiv1.CustomResourceDefinition{}

	err = k8sClient.Get(ctx, types.NamespacedName{Name: "meshsyncs.meshplay.khulnasoft.com"}, crd)
	Expect(err).NotTo(HaveOccurred())
	err = k8sClient.Get(ctx, types.NamespacedName{Name: "brokers.meshplay.khulnasoft.com"}, crd)
	Expect(err).NotTo(HaveOccurred())
	Expect(crd.Spec.Names.Kind).To(Equal("Broker"))
})

var _ = AfterSuite(func() {
	err := testEnv.Stop()
	if err != nil {
		os.Exit(1)
	}
})