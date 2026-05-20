function updateTheme() {
  var t = localStorage.getItem("ac_theme") || "auto";
  if (t === "dark" || (t === "auto" && window.matchMedia("(prefers-color-scheme: dark)").matches)) {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
}

function setTheme(t) {
  localStorage.setItem("ac_theme", t);
  updateTheme();
}

function setLanguage(l) {
  if (l === "") {
    document.cookie = "user_language=; path=/; max-age=-1";
  } else {
    document.cookie = "user_language=" + l + "; path=/; max-age=31536000";
  }
  location.reload();
}

(function () {
  updateTheme();
})();
