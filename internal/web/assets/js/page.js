function updateTheme() {
  var t = localStorage.getItem("ac_theme") || "auto";
  if (t === "dark" || (t === "auto" && window.matchMedia("(prefers-color-scheme: dark)").matches)) {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
  var options = document.querySelectorAll("[data-theme-option]");
  for (var i = 0; i < options.length; i++) {
    options[i].setAttribute(
      "aria-pressed",
      options[i].getAttribute("data-theme-option") === t ? "true" : "false",
    );
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

function submitForm(id) {
  var form = document.getElementById(id);
  if (form) {
    form.requestSubmit();
  }
}

(function () {
  updateTheme();
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", updateTheme);
  }
})();
