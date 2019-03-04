import React, { Component } from 'react';
import {
    Collapse,
    Navbar,
    NavbarToggler,
    NavbarBrand,
    Nav,
    NavItem,
    NavLink,
    Container,
    Row,
    Col,
    Jumbotron,
    Table,
    Badge,
    Button,
    Form,
    FormGroup,
    Label,
    Input,
    InputGroup,
    InputGroupAddon
} from 'reactstrap';
import {BigNumber} from 'bignumber.js';
import QRCode from 'qrcode.react';

class App extends Component {
    constructor(props) {
        super(props);
        this.baseUrl = "/api/"
        if(window.location.host == "localhost:3000") { // When running in REACT Dev
            this.baseUrl = "http://localhost:27888/api/"
        }
        this.toggle = this.toggle.bind(this);
        this.state = {
            isOpen: false,
            balance: new BigNumber(0),
            page: 'home',
            vertcoinAddress: "",
            assetAddress: "",
            assets: [],
            network: "",
            sendAsset: null,
            sendAmount: "0",
            sendTo:"",
            issueTicker: "",
            issueDecimals: 8,
            issueSupply: 0,
            donating: false,
        };
        this.refreshAssets = this.refreshAssets.bind(this);
        this.refreshBalance = this.refreshBalance.bind(this);
        this.sendAsset = this.sendAsset.bind(this);
        this.issueAsset = this.issueAsset.bind(this);
        this.refreshNetwork = this.refreshNetwork.bind(this);
        this.refreshAddresses = this.refreshAddresses.bind(this);
        this.refresh = this.refresh.bind(this);
    }

    sendAsset(asset, amount, to) {
        var realAmount = new BigNumber(amount).times(new BigNumber("1e" + asset.Decimals.toString())).toNumber();
        fetch(this.baseUrl + "transferAsset", {
            method: "POST",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                
                    "AssetID": asset.AssetID,
                    "Amount" : realAmount,
                    "RecipientAddress": to
                
            })
        })
        .then((res) => { return res.json(); })
        .then((data) => {
            this.setState({page:'home'});
        })
    }

    issueAsset(ticker, decimals, supply) {
        var realAmount = new BigNumber(supply).times(new BigNumber("1e" + decimals.toString())).toNumber();
        fetch(this.baseUrl + "newAsset", {
            method: "POST",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                    "Ticker": ticker,
                    "TotalSupply" : realAmount,
                    "Decimals": parseInt(decimals)
            })
        })
        .then((res) => { return res.json(); })
        .then((data) => {
            this.setState({page:'home'});
        })
    }

    refreshBalance() {
        fetch(this.baseUrl + "balance").then((resp) => resp.json()).then((resp) => {
            this.setState({
                balance: new BigNumber(resp.TotalBalance)
            });
        });
    }
    refreshAssets() {
        fetch(this.baseUrl + "assets/mine").then((resp) => resp.json()).then((resp) => {
            if(resp.Assets === null) { resp.Assets = []; }
            this.setState({
                assets : resp.Assets
            });
        });
    }
    refreshNetwork() {
        fetch(this.baseUrl + "network").then((resp) => resp.json()).then((resp) => {
            this.setState({
                network : resp.NetworkName
            });
        });
    }
    
    refreshAddresses() {
        fetch(this.baseUrl + "addresses").then((resp) => resp.json()).then((resp) => {
            this.setState({
                vertcoinAddress: resp.VertcoinAddress,
                assetAddress: resp.AssetAddress
            })
        });
    }

    componentDidMount() {
        this.refreshNetwork();
        this.refresh();
        this.refreshInterval = setInterval(this.refresh, 5000);
    }

    componentWillUnmount() {
        clearInterval(this.refreshInterval);
    }

    refresh() {
        this.refreshBalance();
        this.refreshAddresses();
        this.refreshAssets();
        
    }

    toggle() {
        this.setState({
            isOpen: !this.state.isOpen
        });
    }
    render() {

        var fractions = this.state.balance.mod(new BigNumber("1e8"))
        var coins = this.state.balance.plus(fractions.negated()).div(new BigNumber("1e8"))
        
        var mainPage = "";
        switch(this.state.page) {
            case 'home':
                var assets = this.state.assets.map((asset) => {
                    var assetBalance = new BigNumber(asset.Balance);
                    var divider = new BigNumber("1e" + asset.Decimals.toString());
                    var assetFractions = assetBalance.mod(divider);
                    var assetAmount = assetBalance.plus(assetFractions.negated()).div(divider);
        
                    return (<tr>
                        <td>{asset.AssetID.substring(0,8)}</td>
                        <td>{asset.Ticker}</td>
                        <td><b>{assetAmount.toString()}</b>.<small>{assetFractions.toString()}</small></td>
                        <td>
                            <Button color="primary" size="sm" onClick={(e) => {
                                this.setState({sendAsset:asset, page:'send'});
                            }}>
                                Send
                            </Button>
                        </td>
                    </tr>);
                });

                mainPage = (<Container>
                    <Row>
                        <Col>
                            <h2>Assets</h2>
                            <Table>
                                <thead>
                                    <tr>
                                        <th>Asset ID</th>
                                        <th>Symbol</th>
                                        <th>Balance</th>
                                        <th>&nbsp;</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {assets}
                                </tbody>
                            </Table>
                        </Col>
                    </Row>
                </Container>)


                break;
            case 'receive':
                mainPage = (<Container>
                    <Row>
                        <Col xs={6} style={{textAlign: "center"}}>
                            <h3>Receive VTC:</h3>
                            <QRCode renderAs="svg" bgColor="rgba(0,0,0,0)" value={this.state.vertcoinAddress} /><br/>
                            <small>{this.state.vertcoinAddress}</small>                        
                        </Col>
                        <Col xs={6} style={{textAlign: "center"}}>
                            <h3>Receive Assets:</h3>
                            <QRCode renderAs="svg" bgColor="rgba(0,0,0,0)" value={this.state.assetAddress} /><br/>
                            <small>{this.state.assetAddress}</small>                        
                        </Col>
                    </Row>
                </Container>);
                break;
            case 'send':
                mainPage = (<Container>
                    <Row>
                        <Col>
                            <Form>
                                <FormGroup row>
                                    <Label for="amount" sm={2}>Amount:</Label>
                                    
                                    <Col sm={10}>
                                        <InputGroup>
                                            <Input type="number" name="amount" id="amount" placeholder="Enter amount" value={this.state.sendAmount} onChange={e => this.setState({ sendAmount: e.target.value })} />
                                            <InputGroupAddon addonType="append">{this.state.sendAsset.Ticker}</InputGroupAddon>
                                        </InputGroup>
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Label for="recipient" sm={2}>Recipient:</Label>
                                    <Col sm={10}>
                                        <Input type="text" name="recipient" id="recipient" placeholder="Enter recipient address" value={this.state.sendTo} onChange={e => this.setState({ sendTo: e.target.value })} />
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Col>
                                        <Button onClick={(e) => {
                                            this.sendAsset(this.state.sendAsset, this.state.sendAmount, this.state.sendTo)
                                        }}>Send</Button>
                                    </Col>
                                </FormGroup>
                            </Form>
                        </Col>
                    </Row>
                </Container>);
                break;
            case 'issue':
                mainPage = (<Container>
                    <Row>
                        <Col>
                            <h2>Issue a new asset</h2>
                            <Form>
                                <FormGroup row>
                                    <Label for="ticker" sm={3}>Ticker:</Label>
                                    <Col sm={9}>
                                        <Input type="text" name="ticker" id="ticker" placeholder="Enter ticker symbol" value={this.state.issueTicker} onChange={e => this.setState({ issueTicker: e.target.value.toUpperCase() })} />
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Label for="amount" sm={3}>Decimals:</Label>
                                    <Col sm={9}>
                                        <Input type="number" name="decimals" id="decimals" placeholder="Enter decimals of the smallest fraction" value={this.state.issueDecimals} onChange={e => this.setState({ issueDecimals: e.target.value })} />
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Label for="supply" sm={3}>Total Supply:</Label>
                                    <Col sm={9}>
                                        <InputGroup>
                                            <Input type="text" name="supply" id="supply" placeholder="Enter total supply to mint (in whole coins)" value={this.state.issueSupply} onChange={e => this.setState({ issueSupply: e.target.value })} />
                                            <InputGroupAddon addonType="append">{this.state.issueTicker}</InputGroupAddon>
                                        </InputGroup>
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Col>
                                        <Button onClick={(e) => {
                                            this.issueAsset(this.state.issueTicker, this.state.issueDecimals, this.state.issueSupply)
                                        }}>Issue {this.state.issueTicker}</Button>
                                    </Col>
                                </FormGroup>
                            </Form>
                        </Col>
                    </Row>
                </Container>);
                break;
        }

        var networkBadge = "";
        if(this.state.network !== "MAINNET") {
            networkBadge = (<Badge color="danger">{this.state.network}</Badge>);
        }
        
        return (
            <div>
                <Navbar color="inverse" light expand="md">
                    <NavbarBrand href="#"  onClick={(e) => { this.setState({page:'home'}) }}><img src="logo.svg" /> Extras {networkBadge}</NavbarBrand>
                    <NavbarToggler onClick={this.toggle} />
                    <Collapse isOpen={this.state.isOpen} navbar>
                        <Nav className="ml-auto" navbar>
                            <NavItem>
                                <NavLink href="#" onClick={(e) => { this.setState({page:'home'}) }}>Home</NavLink>
                            </NavItem>
                            <NavItem>
                                <NavLink href="#" onClick={(e) => { this.setState({page:'receive'}) }}>Receive</NavLink>
                            </NavItem>
                            <NavItem>
                                <NavLink href="#" onClick={(e) => { this.setState({page:'issue'}) }}>Issue</NavLink>
                            </NavItem>
                            <NavItem>
                                <div class="nav-link">
                                    | Balance: <b>{ coins.toString() }</b>.<small>{ fractions.toString() }</small>&nbsp;<b>VTC</b>
                                </div>
                            </NavItem>
                        </Nav>
                    </Collapse>
                </Navbar>
                <Jumbotron>
                    <Container>
                        <Row>
                            <Col>
                                {mainPage}
                            </Col>
                        </Row>
                    </Container>
                </Jumbotron>
            </div>
        );
    }
}

export default App;