import FeedBoard from "./FeedBoard";
import Divider from "./Divider";
import Footer from "./Footer";
import CardCollection from "./CardCollection";
import { useState } from "react";
import Navbar, { NavbarButton } from "./Navbar";
import { cardCollections } from "./CardCollections";

function App() {
    const [show, setShow] = useState(false);

    return (
        <>
            <Navbar
                buttonsLeft={cardCollections.map((cc) => {
                    return (
                        <NavbarButton
                            value={cc.title ?? ""}
                            url={`#${cc.title}`}
                        />
                    );
                })}
            />
            {cardCollections.map((cc) => {
                return <CardCollection title={cc.title} cards={cc.cards} />;
            })}
            <Divider />
            <button onClick={() => setShow(!show)}>Show feedboard</button>
            {show ? <FeedBoard /> : <></>}
            <Footer />
        </>
    );
}

export default App;
