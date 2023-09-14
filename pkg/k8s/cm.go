/*
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-05-09 19:49:35
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-09-14 18:05:14
 * @FilePath: /Users/zengshenglong/Code/GoWorkSpace/operators/mysql-operator/pkg/k8s/cm.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */

package k8s

import (
	"bytes"
	"context"
	"fmt"

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

	// ex, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }
	// exPath := filepath.Dir(ex)
	// realPath, err := filepath.EvalSymlinks(exPath)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(filepath.Dir(realPath))

	// cfg, err := ini.Load(fmt.Sprintf("%s/cfg/my.cnf", realPath))
	cfg := ini.Empty()
	mysqldSection, err := cfg.NewSection("mysqld")
	mysqldSection.NewBooleanKey("log_slave_updates")
	mysqldSection.NewBooleanKey("!includedir /etc/mysql/mysql.conf.d/")
	data, _ := writeConfigs(cfg)
	err = c.Get(ctx, ctrl.NamespacedName, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Create ConfigMap")
			cm.Data = map[string]string{
				"my.cnf": data,
			}
			c.Create(ctx, cm)
			return nil
		}
		fmt.Println("error creating ConfigMap")
	}

	cm.Data = map[string]string{
		"my.cnf": data,
	}
	err = c.Update(ctx, cm)
	if err != nil {
		log.Info("Update ConfigMap failed")
		return err
	}
	fmt.Print("Updated ConfigMap successfully")
	return nil
}

func writeConfigs(cfg *ini.File) (string, error) {
	var buf bytes.Buffer
	if _, err := cfg.WriteTo(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
