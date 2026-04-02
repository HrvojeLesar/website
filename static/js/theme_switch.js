function themeSwitch() {
    const buttons = document.getElementsByClassName("theme-switch");
    let isDark = isDarkModeEnabled();
    setDocumentTheme(isDark);
    for (const button of buttons) {
        button.addEventListener("click", () => {
            toggleLightMode();
            isDark = isDarkModeEnabled();
            setDocumentTheme(isDark);
        });
    }

    window
        .matchMedia("(prefers-color-scheme: light)")
        .addEventListener("change", (e) => {
            if (localStorage.getItem("light-mode") === null) {
                setDocumentTheme(!e.matches);
            }
        });
}

function setDocumentTheme(isDark) {
    document.documentElement.setAttribute("class", isDark ? "dark" : "light");
}

function isDarkModeEnabled() {
    return !isLightModeEnabled();
}

function isLightModeEnabled() {
    const stored = localStorage.getItem("light-mode");
    if (stored !== null) {
        return stored === "true";
    }

    return window.matchMedia("(prefers-color-scheme: light)").matches;
}

function toggleLightMode() {
    localStorage.setItem("light-mode", isLightModeEnabled() ? "false" : "true");
}

themeSwitch();
