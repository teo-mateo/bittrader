import React from 'react'
import axios from 'axios'

class TradePairs extends React.Component{
    constructor(){
        super();
        this.state = {};
    }


    componentDidMount(){
        axios.get('https://api.kraken.com/0/public/AssetPairs')
            .then(function(data){
                console.log('axios got stuff: ', data);
            })
            .catch(function(error){
                console.log(error);
            })
    }

    render(){
        return (
            <div> Trade pairs: </div>
        )
    }
}

module.exports = TradePairs;