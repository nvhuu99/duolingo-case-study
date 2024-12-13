const { watch, series } = require('gulp');
const { exec, spawn } = require('child_process');

function goBuild(cb) {
  console.log("Building server...");
  exec(`cd "$APP_SRC" && go build -gcflags "all=-N -l" -o "$APP_DIR/bin/server" ./migrate.go`, (err) => {
    cb();
  });
}

function startDebug(cb) {
  spawn('dlv', [
    '--listen=:4000',
    '--headless=true',
    '--log=true',
    '--accept-multiclient',
    '--api-version=2',
    'exec',
    'bin/server'
  ]);
  console.log(`Debug server started`);
  cb();
}

function stopDebug(cb) {
  exec('pidof dlv', (err, pId) => {
    pId = pId?.trim();
    if (pId) {
      exec(`kill ${pId}`, () => {
        console.log("Debug server stopped");
        cb();
      });
    }
    else {
      cb();
    }
  })
}

exports.default = function() {
  watch('src/**/*', { usePolling: true }, series(goBuild, stopDebug, startDebug));
};
