import FeedBoard from "./FeedBoard";
import Divider from "./Divider";
import Footer from "./Footer";
import CardCollection from "./CardCollection";
import { eveStuff, projects } from "./ProjectsCollection";
import { useState } from "react";

function App() {
    const [show, setShow] = useState(false);

    return (
        <>
            <CardCollection title="Projects" cards={projects} />
            <CardCollection title="Eve stuff" cards={eveStuff} />
            <Divider />
            <button onClick={() => setShow(!show)}>Show feedboard</button>
            {show ? <FeedBoard /> : <></>}
            <Footer />
        </>
    );
}

export default App;
