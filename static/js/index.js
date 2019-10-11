let dataTables = DataTables.default;

Vue.component('tabel-detail', {
    delimiters: ['{[', ']}'],
    template: '#tabel-detail-template',
    components: {dataTables},
    data: function () {
        return {
            tableData: [],
            gridData: [],
            dialogFormVisible: false,
            scanTableVisible: false,
            form: {
                name: '',
                index: ''
            },
            running: {
                heart_beat: '',
                peace_count: '',
                meter_count: '',
                calorie: '',
                peace_speed: '',
                data: '',
                bracelet_name: ''
            },
            reportTitle: '',
            statusVisible: false,
            formType: 'create',
            formTitle: '添加数据',
            loading: false,
            websocket: null
        }
    },
    mounted: function () {
        this.getCategories();
        this.initWebSocket()
    },
    methods: {
        getActionsDef: function () {
            let self = this;
            return {
                width: 5,
                def: [{
                    name: '扫描设备',
                    handler() {
                        self.$confirm('确认扫描周边蓝牙设备?', '提示', {
                            confirmButtonText: '确定',
                            cancelButtonText: '取消',
                            type: 'warning'
                        }).then(function () {
                            self.scanTableVisible = true
                            self.scan()
                        });
                    },
                    icon: 'plus'
                }]
            }
        },
        getConnectedRowActionDef: function () {
            let self = this;
            return [{
                type: 'primary',
                handler(row) {
                    self.$confirm('确认开始实时监测心率?', '提示', {
                        confirmButtonText: '确定',
                        cancelButtonText: '取消',
                        type: 'warning'
                    }).then(function () {
                        self.sendMessage('start', row.localName)
                    });
                },
                name: '开始监测'
            }, {
                type: 'primary',
                handler(row) {
                    self.$confirm('确认停止实时监测心率?', '提示', {
                        confirmButtonText: '确定',
                        cancelButtonText: '取消',
                        type: 'warning'
                    }).then(function () {
                        self.sendMessage('stop', row.localName)
                    });
                },
                name: '停止监测'
            }]
        },
        getPaginationDef: function () {
            return {
                pageSize: 10,
                pageSizes: [10, 20, 50]
            }
        },
        getRowActionsDef: function () {
            let self = this;
            return [{
                type: 'primary',
                handler(row) {
                    self.$confirm('确认连接该设备?', '提示', {
                        confirmButtonText: '确定',
                        cancelButtonText: '取消',
                        type: 'warning'
                    }).then(function () {
                        // self.reportTitle = "实时数据";
                        // self.statusVisible = true;
                        self.sendMessage('open', row.index)
                        self.scanTableVisible = false


                        // let url = Flask.url_for("connect");
                        // axios.post(url, {name: row.addr, index: row.index}).then(function (response) {
                        //     self.$message.success("连接成功！")
                        //
                        // }).catch(self.showError)
                    });
                },
                name: '连接'
            }]
        },
        getCategories: function () {
            let url = "/get_base_data";

            let self = this;
            self.loading = true
            axios.get(url).then(function (response) {
                self.tableData = response.data;
                self.loading = false
            });
        },
        scan: function () {
            let url = "/scan";
            let self = this;
            self.loading = true
            axios.get(url).then(function (response) {
                self.gridData = response.data;
                self.loading = false
            });

        },
        connect: function () {
            let url = "/connect";
            axios.get(url)
        },
        createOrUpdate: function () {
            let self = this;
            if (self.form.name === '') {
                self.$message.error('数据不能为空！');
                return
            }
            if (self.formType === 'create') {
                let url = "add";
                axios.post(url, {
                    name: self.form.name
                }).then(function (response) {
                    self.getCategories();
                    self.dialogFormVisible = false;
                    self.$message.success('添加成功！')
                }).catch(self.showError);
            } else {
                let url = "update";
                axios.put(url, {
                    name: self.form.name,
                    index: self.form.index
                }).then(function (response) {
                    self.getCategories();
                    self.dialogFormVisible = false;
                    self.$message.success('修改成功！')
                }).catch(self.showError);
            }
        },
        closeBracelet: function () {
            this.sendMessage('close', this.running.bracelet_name);
            this.statusVisible = false;
        },
        showError: function (error) {
            let response = error.response;
            this.$message.error(response.data.message);
        },
        initWebSocket: function () { //初始化weosocket
            //ws地址
            let self = this;

            namespace = '/test';

            wsuri = (location.protocol + '//' + document.domain + ':' + location.port + namespace);
            this.websocket = io.connect(wsuri)
            this.websocket.on('server_response', self.websocketOnMessage)
            this.websocket.on('command_response', self.commandOnMessage)
            this.websocket.onclose = self.websocketClose;
            this.websocket.onopen = self.onOpen;
        },
        websocketOnMessage: function (e) { //数据接收
            console.log(e);
            this.running = e;

        },
        sendMessage: function (commCh, agentData) {//数据发送
            this.websocket.emit(commCh, agentData);
        },
        websocketClose: function (e) {  //关闭
            console.log("connection closed (" + e.code + ")");
        },
        onOpen: function (e) {
            console.log("connection open" + e.data())
        },
        commandOnMessage: function (e) {
            console.log("command_response" + e)

            this.getCategories()

        }

    }

});

new Vue({
    el: '#vue-app'
});




