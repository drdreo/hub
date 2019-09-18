import React, {Component} from "react";
import {withRouter} from "react-router-dom";
import axios from "axios";

import "./Home.scss";

console.log(process.env);
const API_URL = process.env.REACT_APP_API_URL;

class Home extends Component {
    constructor(props) {
        super(props);
        this.state = {
            room: "",
            username: ""
        };

    }

    render() {
        return (
            <div className="page-container">
                <h4>Owe Drahn</h4>
                <div className="form">
                    <input className="input username" value={this.state.username}
                           onChange={evt => this.updateUsername(evt)}
                           placeholder="Username"/>
                    <input className="input room" value={this.state.room} onChange={evt => this.updateRoom(evt)}
                           placeholder="Room"/>
                    <button className="button join" onClick={() => this.joinGame()}>Join</button>
                </div>
            </div>
        );
    }

    updateRoom(evt) {
        this.setState({
            room: evt.target.value
        });
    }

    updateUsername(evt) {
        this.setState({
            username: evt.target.value
        });
    }

    joinGame() {
        const room = this.state.room;
        const username = this.state.username;

        axios.get(`${API_URL}/join?room=${room}&username=${username}`, {withCredentials: true})
            .then((response) => {
                console.log(response);
                sessionStorage.setItem("playerId", response.data.playerId);
                this.props.history.push("/game/" + room);
            });
    }
}

export default withRouter(Home);
