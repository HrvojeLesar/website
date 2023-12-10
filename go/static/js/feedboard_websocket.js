function startConnection() {
    const conn = new WebSocket(`ws://${location.host}/feedboard-subscribe`);

    conn.addEventListener("open", () => {
        console.log("Connected to feedboard websocket");
    });

    conn.addEventListener("close", () => {
        console.log("Feedboard websocket closed");
    });

    conn.addEventListener("message", (event) => {
        if (typeof event.data !== "string") {
            console.error("Unexpected message type");
            return;
        }
        switchFeedboard(event.data);
    });
}

const parser = new DOMParser();
const feedboard = document.getElementsByClassName("feedboard-kills-container")[0];

function switchFeedboard(newFeedboard) {
    const newFeedboardElement = parser.parseFromString(newFeedboard, "text/html").body.firstChild;
    const lastChildElement = feedboard.children[feedboard.children.length - 1];
    feedboard.removeChild(lastChildElement);

    const firstChild = feedboard.firstChild;
    if (firstChild) {
        feedboard.insertBefore(newFeedboardElement, firstChild);
    } else {
        feedboard.appendChild(newFeedboardElement);
    }
}

if (feedboard) {
    startConnection();
}
