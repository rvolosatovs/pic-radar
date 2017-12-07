import React, { Component } from 'react';
import { FormGroup, FormControl } from 'react-bootstrap';

export class SearchBar extends Component {
    constructor(props) {
        super(props);
        this.handleInput = this.handleInput.bind(this);
        this.state = {
            input: '',
            timer: null
        };
        this.cb = this.props.cb;
    }

    handleInput(e) {
        if (this.state.timer != null) {
            clearTimeout(this.state.timer);
        }

        let v = e.target.value;
        this.setState({
            input: v,
            timer: setTimeout(() => {
                if (this.cb !== null) {
                    this.cb(v);
                }
            }, 50)
        });
    }

    render() {
        return (
            <FormGroup bsSize="large">
                <FormControl value={this.state.input} onChange={this.handleInput} type="text" placeholder="Input a search query here..." />
            </FormGroup>
        )}
}
