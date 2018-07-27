function StatisticsConn(url, options) {
    options = options || {};
    var uri = 'ws://' + url;
    if (window.location.protocol === 'https:') {
        uri = 'wss://' + url;
    }
    var opts = {
        version: 1,
        uri: uri,
        data_opts: { type: "application/octet-stream" },
        status: 0,
        keepalive: 5000,
        id: options.id,
        token: options.token,
        onopen: options.onopen,
        onconnected: options.onconnected,
        onclose: options.onclose,
        onheartbeat: options.onheartbeat,
    }
    var ws = new WebSocket(uri);
    ws.onopen = function () {
        connect();
        if (opts.onopen) {
            opts.onopen()
        }
    }
    ws.onclose = function () {
        opts.status = 0
        if (opts.onclose) {
            opts.onclose()
        }
    }
    ws.onmessage = function (evt) {
        var pkg = readPkg(evt.data, function (pkg) {
            if (!pkg) {
                console.log("StatisticsConn read package error")
                return
            }
            if (pkg.type == 2) {
                if (opts.status == 0) {
                    opts.status = 1;
                    opts.keepTimer = setInterval(function () {
                        if (opts.status == 1) {
                            var playload = pingPkg();
                            ws.send(playload);
                        } else {
                            clearInterval(opts.keepTimer);
                        }
                    }, opts.keepalive);
                }
                if (opts.onconnected) {
                    opts.onconnected(pkg)
                }
            } else if (pkg.type == 4) {
                if (opts.onheartbeat) {
                    opts.onheartbeat()
                }
            }
        })
    }
    function connect() {
        var playload = connectPkg(opts.id, opts.token, window.location.host);
        ws.send(playload);
    }

    function readPkg(blob, recvFn) {
        var reader = new FileReader();
        reader.readAsArrayBuffer(blob);
        reader.onload = function (e) {
            var pkg = {};
            var view = new DataView(reader.result);
            pkg.version = view.getUint16(0, false);
            pkg.type = view.getUint8(2)
            pkg.flags = view.getUint8(3)
            pkg.length = view.getUint32(4, false)
            if (pkg.length > 0) {
                reader.readAsText(blob.slice(8), 'utf-8');
                reader.onload = function () {
                    pkg.data = reader.result;
                    recvFn(pkg)
                };
            } else {
                recvFn(pkg)
            }
        }
    }
    function connectPkg(id, token, domain) {
        var buffer = new ArrayBuffer(8);
        var view = new DataView(buffer);
        view.setUint16(0, opts.version, false);
        view.setUint8(2, 1, false);
        view.setUint8(3, 0, false);
        var data = JSON.stringify({ id: id, token: token, domain: domain });
        view.setUint32(4, data.length, false);
        return new Blob([buffer, data], opts.data_opts);
    }
    function pingPkg() {
        var buffer = new ArrayBuffer(8);
        var view = new DataView(buffer);
        view.setUint16(0, opts.version, false);
        view.setUint8(2, 3, false);
        view.setUint8(3, 0, false);
        view.setUint32(4, 0, false);
        return new Blob([buffer], opts.data_opts);
    }
}