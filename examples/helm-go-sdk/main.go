package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// HelmClient 封装 Helm 操作
type HelmClient struct {
	settings   *cli.EnvSettings
	actionConfig *action.Configuration
}

// NewHelmClient 创建 Helm 客户端
func NewHelmClient(namespace string) (*HelmClient, error) {
	settings := cli.New()

	actionConfig := new(action.Configuration)

	// 初始化 Helm action configuration
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		namespace,
		os.Getenv("HELM_DRIVER"), // 可以是 "configmap", "secret", "memory"
		func(format string, v ...interface{}) {
			log.Printf(format, v...)
		},
	); err != nil {
		return nil, err
	}

	return &HelmClient{
		settings:     settings,
		actionConfig: actionConfig,
	}, nil
}

// InstallChart 安装 Helm Chart
func (h *HelmClient) InstallChart(releaseName, chartPath string, values map[string]interface{}) (*release.Release, error) {
	client := action.NewInstall(h.actionConfig)
	client.Namespace = h.actionConfig.Namespace
	client.ReleaseName = releaseName
	client.CreateNamespace = true
	client.Wait = true
	client.Timeout = 300 * 1000000000 // 5 minutes in nanoseconds

	// 加载 Chart
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	// 安装
	rel, err := client.Run(chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to install chart: %w", err)
	}

	return rel, nil
}

// UpgradeChart 升级 Helm Chart
func (h *HelmClient) UpgradeChart(releaseName, chartPath string, values map[string]interface{}) (*release.Release, error) {
	client := action.NewUpgrade(h.actionConfig)
	client.Namespace = h.actionConfig.Namespace
	client.Wait = true
	client.Timeout = 300 * 1000000000

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	rel, err := client.Run(releaseName, chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade chart: %w", err)
	}

	return rel, nil
}

// UninstallChart 卸载 Helm Release
func (h *HelmClient) UninstallChart(releaseName string) error {
	client := action.NewUninstall(h.actionConfig)

	_, err := client.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall chart: %w", err)
	}

	return nil
}

// ListReleases 列出所有 releases
func (h *HelmClient) ListReleases() ([]*release.Release, error) {
	client := action.NewList(h.actionConfig)
	client.All = true

	releases, err := client.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}

	return releases, nil
}

// GetRelease 获取指定 release 的信息
func (h *HelmClient) GetRelease(releaseName string) (*release.Release, error) {
	client := action.NewGet(h.actionConfig)

	rel, err := client.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	return rel, nil
}

// RollbackRelease 回滚 release 到指定版本
func (h *HelmClient) RollbackRelease(releaseName string, version int) error {
	client := action.NewRollback(h.actionConfig)
	client.Version = version
	client.Wait = true

	if err := client.Run(releaseName); err != nil {
		return fmt.Errorf("failed to rollback release: %w", err)
	}

	return nil
}

// GetReleaseHistory 获取 release 历史
func (h *HelmClient) GetReleaseHistory(releaseName string) ([]*release.Release, error) {
	client := action.NewHistory(h.actionConfig)
	client.Max = 10 // 最多返回 10 个历史版本

	history, err := client.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get release history: %w", err)
	}

	return history, nil
}

// GetReleaseValues 获取 release 的 values
func (h *HelmClient) GetReleaseValues(releaseName string) (map[string]interface{}, error) {
	client := action.NewGetValues(h.actionConfig)

	values, err := client.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get release values: %w", err)
	}

	return values, nil
}

func main() {
	// 示例: 使用 Helm Go SDK

	namespace := "es-serverless"

	// 创建 Helm 客户端
	helmClient, err := NewHelmClient(namespace)
	if err != nil {
		log.Fatalf("Failed to create helm client: %v", err)
	}

	// 1. 安装 Elasticsearch Chart
	fmt.Println("Installing Elasticsearch chart...")
	values := map[string]interface{}{
		"replicaCount": 3,
		"persistence": map[string]interface{}{
			"enabled":      true,
			"storageClass": "hostpath",
			"size":         "10Gi",
		},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    "1000m",
				"memory": "2Gi",
			},
		},
	}

	rel, err := helmClient.InstallChart(
		"elasticsearch",
		"../../helm/elasticsearch",
		values,
	)
	if err != nil {
		log.Fatalf("Failed to install: %v", err)
	}
	fmt.Printf("Installed release: %s (version %d)\n", rel.Name, rel.Version)

	// 2. 列出所有 releases
	fmt.Println("\nListing all releases...")
	releases, err := helmClient.ListReleases()
	if err != nil {
		log.Fatalf("Failed to list releases: %v", err)
	}
	for _, r := range releases {
		fmt.Printf("  - %s (namespace: %s, status: %s, version: %d)\n",
			r.Name, r.Namespace, r.Info.Status, r.Version)
	}

	// 3. 获取 release 信息
	fmt.Println("\nGetting release info...")
	release, err := helmClient.GetRelease("elasticsearch")
	if err != nil {
		log.Fatalf("Failed to get release: %v", err)
	}
	fmt.Printf("Release: %s\n", release.Name)
	fmt.Printf("Chart: %s-%s\n", release.Chart.Name(), release.Chart.Metadata.Version)
	fmt.Printf("Status: %s\n", release.Info.Status)

	// 4. 升级 release
	fmt.Println("\nUpgrading release...")
	newValues := map[string]interface{}{
		"replicaCount": 5, // 增加副本数
	}
	upgradedRel, err := helmClient.UpgradeChart(
		"elasticsearch",
		"../../helm/elasticsearch",
		newValues,
	)
	if err != nil {
		log.Fatalf("Failed to upgrade: %v", err)
	}
	fmt.Printf("Upgraded to version %d\n", upgradedRel.Version)

	// 5. 获取 release values
	fmt.Println("\nGetting release values...")
	releaseValues, err := helmClient.GetReleaseValues("elasticsearch")
	if err != nil {
		log.Fatalf("Failed to get values: %v", err)
	}
	fmt.Printf("Current values: %+v\n", releaseValues)

	// 6. 获取历史
	fmt.Println("\nGetting release history...")
	history, err := helmClient.GetReleaseHistory("elasticsearch")
	if err != nil {
		log.Fatalf("Failed to get history: %v", err)
	}
	for _, h := range history {
		fmt.Printf("  Version %d: %s (deployed at %s)\n",
			h.Version, h.Info.Status, h.Info.FirstDeployed)
	}

	// 7. 回滚 (可选)
	// fmt.Println("\nRolling back to version 1...")
	// err = helmClient.RollbackRelease("elasticsearch", 1)
	// if err != nil {
	// 	log.Fatalf("Failed to rollback: %v", err)
	// }

	// 8. 卸载 (可选)
	// fmt.Println("\nUninstalling release...")
	// err = helmClient.UninstallChart("elasticsearch")
	// if err != nil {
	// 	log.Fatalf("Failed to uninstall: %v", err)
	// }
	// fmt.Println("Uninstalled successfully")
}
