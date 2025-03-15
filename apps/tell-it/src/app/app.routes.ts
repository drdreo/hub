import { Routes } from "@angular/router";

export const appRoutes: Routes = [
    {
        path: "",
        loadComponent: () => import("@tell-it-web/home").then(m => m.HomeComponent)
    },
    {
        path: "room/:roomName",
        loadComponent: () => import("@tell-it-web/room").then(mod => mod.RoomComponent)
    }
];
