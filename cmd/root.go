// Copyright © 2018 RecallSong <songruiguo@qq.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"syscall"
	"time"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/os/signalx"
	"github.com/recallsong/online-statistics/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() {
	cobrax.Execute("online-statistics", options)
}

var (
	svrCfg  server.Config
	options = &cobrax.Options{
		CfgDir:      "conf",
		CfgFileName: "config",
		AppConfig:   &svrCfg,
		Init: func(cmd *cobra.Command) {
			cmd.Short = "tcp connects statistics"
			cmd.Long = `tcp and websocket connects statistics.`
			fs := cmd.Flags()
			fs.StringVar(&svrCfg.TcpAddr, "tcp_addr", "", "tcp server listen at this address")
			fs.StringVar(&svrCfg.TcpTLSAddr, "tcp_tls_addr", "", "tcp (tls) server listen at this address")
			fs.StringVar(&svrCfg.HttpAddr, "http_addr", "", "http server listen at this address")
			fs.StringVar(&svrCfg.HttpsAddr, "https_addr", "", "https server listen at this address")
			fs.StringVar(&svrCfg.AdminAddr, "admin_addr", "", "admin http server listen at this address")
			fs.DurationVar(&svrCfg.KeepAlive, "keepalive", 5*time.Second, "keepalive for connect read timeout")
			fs.StringVar(&svrCfg.ConnCheckUrl, "conn_check_url", "", "url for check connect")
			viper.BindPFlags(fs)
		},
		Run: func(cmd *cobra.Command, args []string) {
			svr := server.New(&svrCfg)
			svr.Start(signalx.Notify(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT))
		},
	}
)
