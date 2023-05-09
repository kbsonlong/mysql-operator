/*
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-05-09 11:17:16
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-05-09 16:59:06
 * @FilePath: /mysql-operator/pkg/k8s/cm.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
/*
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-05-09 11:17:16
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-05-09 16:58:43
 * @FilePath: /mysql-operator/pkg/k8s/cm.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package k8s

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	batchv1 "github.com/kbsonlong/mysql-operator/api/v1"
	"gopkg.in/ini.v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CreateConfig(c client.Client, ctx context.Context, ctrl ctrl.Request, cluster *batchv1.MysqlCluster) error {
	log := log.FromContext(ctx)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}
	err := c.Get(ctx, ctrl.NamespacedName, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Create ConfigMap")
			ex, err := os.Executable()
			if err != nil {
				panic(err)
			}
			exPath := filepath.Dir(ex)
			realPath, err := filepath.EvalSymlinks(exPath)
			if err != nil {
				panic(err)
			}
			fmt.Println(filepath.Dir(realPath))

			cfg, err := ini.Load(fmt.Sprintf("%s/cfg/my.cnf", realPath))
			if err != nil {
				fmt.Println("new mysql section failed:", err)
				return err
			}
			fmt.Println(cfg.MapTo(map[string]string{}))
			data, _ := writeConfigs(cfg)
			cm.Data = map[string]string{
				"my.cnf": data,
			}

			c.Create(ctx, cm)
		}
	}
	return nil
}

func writeConfigs(cfg *ini.File) (string, error) {
	var buf bytes.Buffer
	if _, err := cfg.WriteTo(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
