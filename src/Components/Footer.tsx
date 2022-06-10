import { contactInfo } from "../Data/ContactInfo";
import Divider from "./Divider";

export default function Footer() {
    return (
        <footer className="py-4 px-16 text-dark-50 overflow-x-hidden">
            <Divider />
            <div className="flex items-center justify-between flex-wrap overflow-x-hidden">
                <div>A really amazing footer.</div>
                <div className="flex items-center gap-3 flex-wrap">
                    {contactInfo.map((ci) => {
                        if (ci.type === "E-Mail") {
                            return (
                                <div
                                    key={ci.type}
                                    className="flex items-center gap-1 flex-wrap"
                                >
                                    {ci.icon({ size: 24 })}
                                    {`${ci.contact}`}
                                </div>
                            );
                        } else {
                            return (
                                <a
                                    href={ci.url}
                                    key={ci.type}
                                >
                                <div
                                    className="hover:fill-dark-500"
                                >
                                    {ci.icon({ size: 24 })}
                                </div>
                                </a>
                            );
                        }
                    })}
                </div>
            </div>
        </footer>
    );
}
