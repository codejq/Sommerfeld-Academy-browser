//go:build windows

package main

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
)

//go:embed build/winres/icon.png
var offlineLogo []byte

// retrySeconds is how long the offline screen waits between attempts.
const retrySeconds = 10

// offlineHTML is shown when the site cannot be loaded (no internet, DNS or
// server error). Every retrySeconds — or as soon as Windows reports the
// network back — it checks that the site answers and then reloads the page
// that failed; a new failure brings this screen back.
const offlineHTML = `<!doctype html>
<html lang="ar" dir="rtl"><head><meta charset="utf-8"><title>Sommerfeld Academy</title>
<style>
  html, body { height: 100%; margin: 0; }
  body { display: flex; align-items: center; justify-content: center; background: #14234b;
         color: #fff; font-family: "Segoe UI", Tahoma, sans-serif; text-align: center; }
  img { width: 140px; height: 140px; margin-bottom: 24px; }
  h1 { color: #f2b705; font-size: 28px; margin: 0 0 8px; }
  p { margin: 4px 0; font-size: 17px; opacity: .9; }
  .en { direction: ltr; font-size: 15px; opacity: .7; }
  .count { margin-top: 22px; font-size: 16px; }
  button { margin-top: 18px; padding: 10px 28px; font-size: 16px; border: 0; border-radius: 6px;
           background: #f2b705; color: #14234b; font-weight: 600; cursor: pointer; }
  button:disabled { opacity: .6; cursor: default; }
</style></head>
<body><main>
  <img src="data:image/png;base64,{{LOGO}}" alt="">
  <h1>تعذّر الاتصال بالأكاديمية</h1>
  <p>تحقّق من اتصال الإنترنت. سنحاول مجدداً تلقائياً.</p>
  <p class="en">Can't reach Sommerfeld Academy. Check the internet connection — retrying automatically.</p>
  <div class="count" id="count"></div>
  <button id="retry">إعادة المحاولة الآن · Retry now</button>
</main>
<script>
(function () {
  var target = {{TARGET}}, home = {{HOME}}, wait = {{SECONDS}};
  var left = wait, busy = false;
  var count = document.getElementById('count'), btn = document.getElementById('retry');
  function show() {
    count.textContent = busy ? 'جارٍ المحاولة… · Trying…'
      : 'المحاولة التالية خلال ' + left + ' ثانية · Next try in ' + left + 's';
    btn.disabled = busy;
  }
  function attempt() {
    if (busy) return;
    busy = true; show();
    fetch(home, { mode: 'no-cors', cache: 'no-store' })
      .then(function () { location.replace(target); })
      .catch(function () { busy = false; left = wait; show(); });
  }
  setInterval(function () {
    if (busy) return;
    if (--left <= 0) attempt(); else show();
  }, 1000);
  btn.onclick = attempt;
  window.addEventListener('online', attempt);
  show();
})();
</script></body></html>`

// offlinePage returns the offline screen that retries url.
func offlinePage(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = homeURL
	}
	js := func(s string) string { b, _ := json.Marshal(s); return string(b) }
	return strings.NewReplacer(
		"{{LOGO}}", base64.StdEncoding.EncodeToString(offlineLogo),
		"{{TARGET}}", js(url),
		"{{HOME}}", js(homeURL),
		"{{SECONDS}}", strconv.Itoa(retrySeconds),
	).Replace(offlineHTML)
}
