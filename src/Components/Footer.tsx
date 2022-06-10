import { contactInfo } from "../Data/ContactInfo";
import Divider from "./Divider";

export default function Footer() {
    return (
        <footer className="py-4 px-16 text-dark-50 overflow-x-hidden bg-dark-700">
            <Divider />
            <div className="flex items-center justify-between flex-wrap overflow-x-hidden gap-y-4 pt-1">
                <div>A really amazing footer.</div>
                <div>Made by <div className="inline-block italic font-semibold">Hrvoje Lesar</div></div>
                <div className="flex items-center gap-x-3 flex-wrap">
                    {contactInfo.map((ci) => {
                        if (ci.type === "E-Mail") {
                            return (
                                <div
                                    key={ci.type}
                                    className="flex items-center gap-1 flex-wrap"
                                >
                                    <a href={ci.url} key={ci.type}>
                                        <div className="hover:fill-dark-500">
                                            {ci.icon({ size: 24 })}
                                        </div>
                                    </a>
                                    {`${ci.contact}`}
                                </div>
                            );
                        } else {
                            return (
                                <a href={ci.url} key={ci.type}>
                                    <div className="hover:fill-dark-500">
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
