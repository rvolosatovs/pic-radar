import React, { Component } from 'react';
import './App.css';
import notFound from './not-found.gif';
import { SearchBar } from './component/SearchBar';
import { Grid, Navbar, Button, Row, Image, Col, Modal, FormControl, FormGroup, NavItem, Nav } from 'react-bootstrap';
import FontAwesome from 'react-fontawesome';
import 'font-awesome/css/font-awesome.css';

const rand = n => Math.floor(Math.random()*n);
const host = "http://2id60.win.tue.nl:8085";
//const host = "http://localhost:8085";

class App extends Component<{}> {
    constructor(props) {
        super(props);
        this.updateImage = this.updateImage.bind(this);
        this.randomize = this.randomize.bind(this);
        this.login = this.login.bind(this);
        this.register = this.register.bind(this);
        this.openModal = this.openModal.bind(this);
        this.spawnLoginOverlay = this.spawnLoginOverlay.bind(this);
        this.spawnRegisterOverlay = this.spawnRegisterOverlay.bind(this);
        this.handleLoginInput  = this.handleLoginInput.bind(this);
        this.handlePasswordInput  = this.handlePasswordInput.bind(this);
        this.handleQueryInput  = this.handleQueryInput.bind(this);
        this.closeModal = this.closeModal.bind(this);
        this.state = {
            image: null,
            alt: 'Random image',
            search: 'hello AND world',
            showModal: false,
            isRegister: false,
            login: '',
            password: '',
            prefQuery: ''
        };
    }

    updateImage(search) {
        if (search == null || typeof search !== 'string' ) {
            console.log('Invalid argument passed to updateImage', search);
            return;
        }
        if (search.length < 1) {
            this.setState({
                image: null,
                search: ''
            });
            return
        }

        let q = host +"/image?q="+encodeURI(search);
        fetch(q, {
            method: "GET",
            headers: {
                'Content-Type': 'application/json',
                'User':  this.state.login
            },
        })
        fetch(q)
            .then(resp => {
                if(resp.ok) {
                    return resp.json();
                }
                throw new Error('Failed to send GET request to ' + q + ':' + resp.text());
            })
            .then(data => {
                let images = [];
                data.data.map(d => {
                    if (d.nsfw || d.is_ad) {
                        return false;
                    }
                    if (!d.is_album) {
                        images.push(d);
                    } else {
                        d.images.map(img => images.push(img));
                    }
                    return true;
                })
                this.setState({
                    image: images[rand(images.length)],
                    search: search
                });
            })
            .catch(err => console.log("Error fetching images", err))
    }

    login() {
        let q = host +"/login";
        fetch(q, {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
                'User':  this.state.login
            },
            body: JSON.stringify({
                login: this.state.login,
                password: this.state.password,
                query: this.state.prefQuery
            })
        })
            .then(resp => {
                if(resp.ok) {
                    resp.json().then(data => {
                        this.setState({
                            login: data.login,
                            password: data.password,
                            prefQuery: data.query
                        });
                    })
                } else {
                    resp.text().then(text => {
                        console.log('Failed to send POST request to ' + q + ':' + text);
                        alert(text);
                    });
                }
            })
            .catch(alert)
        this.closeModal()
    }

    register() {
        let q = host +"/register";
        fetch(q, {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
                'User':  this.state.login
            },
            body: JSON.stringify({
                login: this.state.login,
                password: this.state.password,
                query: this.state.prefQuery
            })
        })
            .then(resp => {
                if(resp.ok) {
                    resp.json().then(data => {
                        this.setState({
                            login: data.login,
                            password: data.password,
                            prefQuery: data.query
                        });
                    })
                } else {
                    resp.text().then(text => {
                        console.log('Failed to send POST request to ' + q + ':' + text);
                        alert(text);
                    });
                }
            })
            .catch(alert)
        this.closeModal()
    }

    spawnLoginOverlay() {
        this.setState({ isRegister: false });
        this.openModal()
    }

    spawnRegisterOverlay() {
        this.setState({ isRegister: true });
        this.openModal()
    }

    randomize() {
        this.updateImage(this.state.search);
    }

    componentDidMount() {
        this.updateImage("Hello AND World");
    }

    closeModal(){
        this.setState({ showModal: false });
    }

    openModal(){
        this.setState({ showModal: true });
    }

    handleLoginInput(event) {
        this.setState({login: event.target.value});
    }

    handlePasswordInput(event) {
        this.setState({password: event.target.value});
    }

    handleQueryInput(event) {
        this.setState({query: event.target.value});
    }

    handleSubmit(event) {
        event.preventDefault();
    }

    render() {
        return (
            <div className="App">
                <Navbar className="navbar" inverse fixedTop >
                    <Navbar.Header>
                        <Navbar.Brand>
                            <a href="http://2id60.win.tue.nl/~s141426/">PicRadar</a>
                        </Navbar.Brand>
                    </Navbar.Header>
                    <Nav>
                        <NavItem
                            onClick={this.spawnLoginOverlay}
                            target="_blank">
                            Login!
                        </NavItem>
                        <NavItem
                            onClick={this.spawnRegisterOverlay}
                            target="_blank">
                            Register!
                        </NavItem>
                    </Nav>
                </Navbar>
                <main>
                    <Grid className="main" fluid>
                        <Row>
                            <h4>This app searches <a href="imgur.com">Imgur</a> for random top 100 images corresponding to the query you type and shows a random one. The app initally searches for 'hello AND world'.<br />
                                Requests are executed concurrently as you type.
                            </h4>
                        </Row>
                        <Row className="search-row">
                            <Col xs={0} md={3} lg={4} />
                            <Col className="center-block" xs={12} mdOffset={3} md={6} lgOffset={4} lg={4} >
                                <SearchBar className="text-center image-search-bar" cb={this.updateImage}/>
                            </Col>
                            <Col xs={0} md={3} lg={4} />
                        </Row>
                        <div className="image-btn-row">
                            <Row>
                                <p>
                                    <Button
                                        className="btn btn-rand btn-lg outline"
                                        bsSize="large"
                                        onClick={this.randomize}
                                        target="_blank">
                                        Randomize!
                                    </Button>
                                </p>
                            </Row>
                        </div>
                        <Row className="image-row">
                            <Modal
                                aria-labelledby='modal-label'
                                show={this.state.showModal}
                                onHide={this.closeModal}
                            >
                                <FormGroup bsSize="large">
                                    <FormControl type="text" name="login" placeholder="Login" onChange={this.handleLoginInput}  />
                                    <FormControl type="password" name="password" placeholder="Password" onChange={this.handlePasswordInput} />
                                    {this.state.isRegister && (
                                        <FormControl type="text" name="query" placeholder="Query" onChange={this.handleQueryInput} show={this.isRegister} />
                                    )}
                                    <Button
                                        className="btn btn-rand btn-lg outline center-block"
                                        bsSize="large"
                                        onClick={this.state.isRegister ? this.register : this.login}
                                        target="_blank">
                                        Execute
                                    </Button>
                                </FormGroup>
                            </Modal>
                            <Col xs={12} md={12} lg={12}>
                                <Image className="image-random center-block" href="#" alt={this.state.alt} src={this.state.image == null ? notFound : this.state.image.link} responsive />
                            </Col>
                        </Row>
                        <hr />
                    </Grid>
                </main>
                <div className="footer">
                    <Grid fluid>
                        <p>Made by s141426<br />
                            <a href="https://github.com/"><FontAwesome name='github' /></a>&nbsp;
                            <a href="https://linkedin.com/in/"><FontAwesome name='linkedin' /></a>&nbsp;
                            <a href="https://facebook.com/"><FontAwesome name='facebook' /></a>&nbsp;
                        </p>
                    </Grid>
                </div>
            </div>
        );
    }
}

export default App;
