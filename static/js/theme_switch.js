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
}

function setDocumentTheme(isDark) {
    document.documentElement.setAttribute(
        "class",
        isDark ? "dark" : "light"
    );
}

function isDarkModeEnabled() {
    return !isLightModeEnabled();
}

function isLightModeEnabled() {
    return localStorage.getItem("light-mode") === "true";
}

function toggleLightMode() {
    if (isLightModeEnabled()) {
        localStorage.setItem("light-mode", "false");
    } else {
        localStorage.setItem("light-mode", "true");
    }
}

themeSwitch();
