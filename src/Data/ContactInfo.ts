import { IconType } from "react-icons";
import { FaEnvelope, FaGithub, FaTwitter } from "react-icons/fa";

type ContactInfo = {
    type: string;
    contact?: string;
    url?: string;
    icon: IconType;
};

export const contactInfo: ContactInfo[] = [
    {
        type: "E-Mail",
        contact: "hrvoje.lesar1@hotmail.com",
        url: "mailto://hrvoje.lesar1@hotmail.com",
        icon: FaEnvelope,
    },
    {
        type: "Github",
        url: "https://github.com/HrvojeLesar",
        icon: FaGithub,
    },
    {
        type: "Twitter",
        url: "https://twitter.com/HrvojeLesar",
        icon: FaTwitter,
    },
];
