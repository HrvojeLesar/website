import { useState } from "react";
import CardCollection from "./Components/CardCollection";
import Divider from "./Components/Divider";
import FeedBoard from "./Components/FeedBoard";
import Footer from "./Components/Footer";
import Navbar, { NavbarButton } from "./Components/Navbar";
import { cardCollections } from "./Data/CardCollections";
import { contactInfo } from "./Data/ContactInfo";

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
