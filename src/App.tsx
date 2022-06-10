import CardCollection from "./Components/CardCollection";
import Divider from "./Components/Divider";
import Footer from "./Components/Footer";
import Navbar, { NavbarButton } from "./Components/Navbar";
import { cardCollections } from "./Data/CardCollections";
import { contactInfo } from "./Data/ContactInfo";

function App() {
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
                buttonsRight={contactInfo.map((ci) => {
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
                return (
                    <CardCollection
                        key={cc.title}
                        title={cc.title}
                        cards={cc.cards}
                    />
                );
            })}
            <Footer />
        </>
    );
}

export default App;
