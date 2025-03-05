import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { enableProdMode, importProvidersFrom } from "@angular/core";
import { bootstrapApplication, BrowserModule } from "@angular/platform-browser";
import { provideRouter, Routes } from "@angular/router";
import { API_URL_TOKEN } from "@tell-it/domain/tokens";
import { SocketIoConfig, SocketIoModule } from "ngx-socket-io";
import { AppComponent } from "./app/app.component";

import { environment } from "./environments/environment";

const socketConfig: SocketIoConfig = {
    url: environment.api.socketUrl,
    options: {}
};
const routes: Routes = [
    {
        path: "",
        loadComponent: () => import("tell-it-home").then(m => m.HomeComponent)
    },
    {
        path: "room/:roomName",
        loadComponent: () => import("tell-it-room").then(mod => mod.RoomComponent)
    }
];

if (environment.production) {
    enableProdMode();
}

bootstrapApplication(AppComponent, {
    providers: [
        importProvidersFrom(BrowserModule, SocketIoModule.forRoot(socketConfig)),
        { provide: API_URL_TOKEN, useValue: environment.api.url },
        provideHttpClient(withInterceptorsFromDi()),
        provideRouter(routes)
    ]
}).catch(err => console.error(err));
