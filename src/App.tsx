import FeedBoard from "./FeedBoard";
import Divider from "./Divider";
import Footer from "./Footer";
import CardCollection from "./CardCollection";
import { eveStuff, projects } from "./ProjectsCollection";

function App() {
    return (
        <>
            <CardCollection title="Projects" cards={projects}/>
            <CardCollection title="Eve stuff" cards={eveStuff}/>
            <Divider />
            <FeedBoard />
            <Footer />
        </>
    );
}

export default App;
