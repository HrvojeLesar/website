import { IInfoCard } from "./IInfoCard";
import PepegaPhone from "../Images/Pepegaphone";
import MM from "../Images/mm.png";

import { FaBook, FaGithub, FaGlobeEurope } from "react-icons/fa";

import PATHFINDER from "../Images/pathfinder.png";
import AocLogo from "../Components/AoCLogo";

export const projects: IInfoCard[] = [
    {
        title: "Kekov soundboard v2 - Beta",
        description:
            "Website for controlling a Discord bot that plays uploaded sounds.",
        moreInfo:
            "Website used to control a Discord bot made for playing short sounds. Original version supported only one server, therefore v2 adds support for multiple servers and an open API for making your own frontend.",
        image: {
            src: PepegaPhone(),
            alt: "Kekov soundboard v2",
        },
        link: "https://betabot.hrveklesarov.com",
        linkButtons: [
            {
                title: "Website",
                url: "https://betabot.hrveklesarov.com",
                icon: FaGlobeEurope,
                iconDecsShort: "Website",
            },
            {
                title: "Github repository",
                url: "https://github.com/HrvojeLesar/kekov_soundboard_v2",
                icon: FaGithub,
                iconDecsShort: "Source code",
            },
            {
                title: "Website v1",
                url: "https://bot.hrveklesarov.com",
                icon: FaGlobeEurope,
                iconDecsShort: "V1 website",
            },
        ],
    },
    {
        title: "Undergraduate thesis",
        description:
            "Web application and separate rust library for solving assignment problems using the Hungarian method.",
        note: "Thesis and website is only available in Croatian.",
        image: {
            src: MM,
            alt: "Undergraduate thesis",
        },
        link: "https://hrveklesarov.com/madarska-metoda",
        linkButtons: [
            {
                title: "Thesis website",
                url: "https://hrveklesarov.com/madarska-metoda",
                icon: FaGlobeEurope,
                iconDecsShort: "Thesis website",
            },
            {
                title: "Thesis repository",
                url: "https://repozitorij.foi.unizg.hr/en/islandora/object/foi%3A7048",
                icon: FaBook,
                iconDecsShort: "Thesis pdf",
            },
            {
                title: "Library source code",
                url: "https://github.com/HrvojeLesar/madarska_metoda",
                icon: FaGithub,
                iconDecsShort: "Lib source",
            },
            {
                title: "Website app source code",
                url: "https://github.com/HrvojeLesar/zavrsni_web",
                icon: FaGithub,
                iconDecsShort: "Website source",
            },
        ],
    },
    {
        title: "Mass balance evidention",
        description: "Nekaj o znidaricaj",
        image: { src: undefined, alt: "Mass balance evidention" },
        link: "https://github.com/HrvojeLesar/Mass-balance-evidention",
        linkButtons: [
            {
                title: "Source code",
                url: "https://github.com/HrvojeLesar/Mass-balance-evidention",
                icon: FaGithub,
                iconDecsShort: "Source code",
            },
        ],
    },
];

export const other: IInfoCard[] = [
    {
        title: "Advent of code",
        description:
            "Advent of Code is an Advent calendar of small programming puzzles for a variety of skill sets and skill levels",
        moreInfo: "Below are links to repositories for my solutions.",
        image: {
            src: AocLogo(),
            alt: "Advent of code",
        },
        link: "https://adventofcode.com",
        linkButtons: [
            {
                title: "Advent of Code 2020 - My solutions",
                url: "https://github.com/HrvojeLesar/AoC2020",
                icon: FaGithub,
                iconDecsShort: "AoC 2020",
            },
            {
                title: "Advent of Code 2021 - My solutions",
                url: "https://github.com/HrvojeLesar/AoC2021",
                icon: FaGithub,
                iconDecsShort: "AoC 2021",
            },
        ],
    },
    {
        title: "Pathfinder",
        description: "Pathfinder is a mapping tool for Eve Online.",
        moreInfo:
            "Official Pathfinder website is down and most others requires you to be a part of their alliance or corporation, so I decided to host my own. As is tradition this requires you to be a port of Fifty Fifty Fifty corporation.",
        image: {
            src: PATHFINDER,
            alt: "Pathfinder logo",
        },
        link: "https://pathfinder.hrveklesarov.com",
        linkButtons: [
            {
                title: "Pathfinder website",
                url: "https://pathfinder.hrveklesarov.com",
                icon: FaGlobeEurope,
                iconDecsShort: "Website",
            },
            {
                title: "Pathfinder source repository",
                url: "https://github.com/exodus4d/pathfinder",
                icon: FaGithub,
                iconDecsShort: "Source code",
            },
        ],
    },
];
