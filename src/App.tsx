import FeedBoard from "./FeedBoard";
import Divider from "./Divider";
import Footer from "./Footer";
import CardCollection from "./CardCollection";
import { useState } from "react";
import Navbar, { NavbarButton } from "./Navbar";
import { cardCollections } from "./CardCollections";
import { contactInfo } from "./ContactInfo";

function App() {
    const [show, setShow] = useState(false);

    return (
        <>
            <Navbar
                buttonsLeft={cardCollections.map((cc) => {
                    return (
                        <NavbarButton
                            key={cc.title}
                            value={cc.title ?? "#"}
                            url={`#${cc.title}`}
                        />
                    );
                })}
                buttonsRight={contactInfo.filter((ci) => { return ci.type !== "E-Mail"}).map((ci) => {
                        return (
                        <NavbarButton
                            key={ci.type}
                            value={ci.type}
                            url={ci.url ?? "#"}
                            icon={ci.icon}
                        />
                        );
                    })}
            />
            {cardCollections.map((cc) => {
                return <CardCollection key={cc.title} title={cc.title} cards={cc.cards} />;
            })}
            <Divider />
            <button onClick={() => setShow(!show)}>Show feedboard</button>
            {show ? <FeedBoard /> : <></>}
            <Footer />
        </>
    );
}

export default App;
