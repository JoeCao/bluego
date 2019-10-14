Vue.component('view-detail', {
    delimiters: ['{[', ']}'],
    template: '#view-detail-template',
    data: function () {
        return {
            runnings: [],
            // running: {
            //     heart_beat: '',
            //     peace_count: '',
            //     meter_count: '',
            //     calorie: '',
            //     peace_speed: '',
            //     data: '',
            //     bracelet_name: ''
            // },
            loading: false,
            websocket: null
        }
    },
    mounted: function () {
        this.initWebSocket()
    },
    methods: {
        fillRunning: function (key, data) {
            console.log(key)
            let index = this.runnings.findIndex(x => x.bracelet_name === key);
            console.log(index)
            if (index >= 0) {
                this.runnings[index] = data
            } else {
                this.runnings.push(data)
            }
            // console.log(this.runnings);
            this.$forceUpdate()
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

            wsuri = ('ws://' + document.domain + ':' + location.port);
            this.websocket = io({transports: ['websocket'], upgrade: false});
            this.websocket.on('server_response', self.websocketOnMessage)
            this.websocket.onmessage = self.websocketOnMessage;
            this.websocket.onclose = self.websocketClose;
            this.websocket.onopen = self.onOpen;
        },
        websocketOnMessage: function (e) { //数据接收
            console.log(e);
            // this.running = e;
            this.fillRunning(e.bracelet_name, e)

        },
        sendMessage: function (commCh, agentData) {//数据发送
            this.websocket.emit(commCh, agentData);
        },
        websocketClose: function (e) {  //关闭
            console.log("connection closed (" + e.code + ")");
        },
        onOpen: function (e) {
            console.log("connection open" + e.data())
        }

    }

});

new Vue({
    el: '#vue-app'
});




