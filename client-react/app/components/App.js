import React from 'react'
import Tradepairs from './startingup/tradepairs'



class App extends React.Component{

    constructor(){
        super();
        this.state = {};
    }

    render(){
        return (
            <div>
                <div>this is the app</div>
                <Tradepairs />
            </div>
        );
    }
}

module.exports = App;