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
import {MdRefresh, MdCheckCircle} from 'react-icons/md';
import {BigNumber} from 'bignumber.js';
import QRCode from 'qrcode.react';

class App extends Component {
    constructor(props) {
        super(props);
        this.baseUrl = "/api/"
        if(window.location.host === "localhost:3000") { // When running in REACT Dev
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
            donating: true,
            connected: false,
            synced: false,
            blockHeight: 0,
            headerQueue: 0,
        };
        this.refreshAssets = this.refreshAssets.bind(this);
        this.refreshBalance = this.refreshBalance.bind(this);
        this.sendAsset = this.sendAsset.bind(this);
        this.issueAsset = this.issueAsset.bind(this);
        this.refreshNetwork = this.refreshNetwork.bind(this);
        this.refreshAddresses = this.refreshAddresses.bind(this);
        this.refreshSyncStatus = this.refreshSyncStatus.bind(this);
        this.refresh = this.refresh.bind(this);
        this.toggleDonations = this.toggleDonations.bind(this);
        this.refreshDonations = this.refreshDonations.bind(this);
    }
    
    toggleDonations() {
        if(this.state.donating === true) {
            fetch(this.baseUrl + "donations/disable")
            .then((res) => { return res.json(); })
            .then((data) => {
                this.setState({donating: false});
            })
        } else {
            fetch(this.baseUrl + "donations/enable")
            .then((res) => { return res.json(); })
            .then((data) => {
                this.setState({donating: true});
            })
        }
    }

    refreshDonations() { 
        fetch(this.baseUrl + "donations/status")
        .then((res) => { return res.json(); })
        .then((data) => {
            this.setState({donating: data});
        })
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

    refreshSyncStatus() {
        fetch(this.baseUrl + "syncStatus").then((resp) => resp.json()).then((resp) => {
            this.setState({
                connected: resp.Connected,
                blockHeight: resp.SyncHeight,
                synced: resp.Synced,
                headerQueue: resp.HeaderQueue,
            }, ()=> {
                if(!(this.state.connected === true) && this.state.newRpcServer === undefined) {
                    fetch(this.baseUrl + "rpcSettings").then((resp) => resp.json()).then((resp) => {
                        this.setState({
                            newRpcServer: resp.RpcHost,
                            newRpcUser: resp.RpcUser,
                            newRpcPassword: resp.RpcPassword
                        });
                    });
                }
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
        this.refreshDonations();
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
        this.refreshSyncStatus()
    }

    updateRpcSettings() {
        fetch(this.baseUrl + "updateRpcSettings", {
            method: "POST",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                
                    "RpcHost": this.state.newRpcServer,
                    "RpcUser" : this.state.newRpcUser,
                    "RpcPassword": this.state.newRpcPassword
                
            })
        })
        .then((res) => { return res.json(); })
        .then((data) => {
            this.setState({connected:true, page:'home'}, () => {this.refresh();});
        })
    }

    toggle() {
        this.setState({
            isOpen: !this.state.isOpen
        });
    }
    render() {
        var page = this.state.page;
        if(!(this.state.connected === true)) {
            page = 'disconnected';
        }
        var fractions = this.state.balance.mod(new BigNumber("1e8"))
        var coins = this.state.balance.plus(fractions.negated()).div(new BigNumber("1e8"))
        
        var mainPage = "";
        switch(page) {
            case 'disconnected':
                mainPage = (<Container>
                <Row>
                    <Col>
                        <h2>Disconnected :-(</h2>
                        <p>Vertcoin OpenAssets is unable to connect to Vertcoin Core. Please check the settings below and correct them if needed:</p>
                        <Form>
                            <FormGroup row>
                                <Label for="amount" sm={2}>RPC URL:</Label>
                                
                                <Col sm={10}>
                                    <InputGroup>
                                        <Input type="text" name="rpcServer" id="rpcServer" placeholder="Enter server URL" value={this.state.newRpcServer} onChange={e => this.setState({ newRpcServer: e.target.value })} />
                                    </InputGroup>
                                </Col>
                            </FormGroup>
                            <FormGroup row>
                                <Label for="amount" sm={2}>RPC User:</Label>
                                
                                <Col sm={10}>
                                    <InputGroup>
                                        <Input type="text" name="rpcUser" id="rpcUser" placeholder="Enter RPC User" value={this.state.newRpcUser} onChange={e => this.setState({ newRpcUser: e.target.value })} />
                                    </InputGroup>
                                </Col>
                            </FormGroup>
                            <FormGroup row>
                                <Label for="amount" sm={2}>RPC Password:</Label>
                                
                                <Col sm={10}>
                                    <InputGroup>
                                        <Input type="text" name="rpcPassword" id="rpcPassword" placeholder="Enter RPC Password" value={this.state.newRpcPassword} onChange={e => this.setState({ newRpcPassword: e.target.value })} />
                                    </InputGroup>
                                </Col>
                            </FormGroup>
                            <FormGroup row>
                                <Col>
                                    <Button onClick={(e) => {
                                        this.updateRpcSettings()
                                    }}>Update</Button>
                                </Col>
                            </FormGroup>
                        </Form>
                    </Col>
                </Row>
                </Container>)
                break;
            default:
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

                var syncStatus = null;
                if(!this.state.synced) {
                    syncStatus = (<Row>
                        <Col>
                            <h2>Catching up...</h2>
                            <p>Vertcoin OpenAssets is catching up with the blockchain. There's <b>{this.state.headerQueue}</b> blocks left to process</p>
                        </Col>
                    </Row>);
                }

                mainPage = (<Container>
                    {syncStatus}
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
            case 'donation':
                var donationButtonIntro = (
                    <p>You are currently not donating. Please consider switching on these tiny donation to support the development of OpenAssets</p>
                );
                if(this.state.donating === true) {
                    donationButtonIntro = (
                        <p>You are currently donating. If want to stop supporting the maintenance of this software, you can switch the donations off.</p>
                    );
                }
                mainPage = (
                    <Container>
                        <Row>
                            <Col>
                                <h2>Donations</h2>
                                <p>The development of Vertcoin and Vertcoin OpenAssets relies entirely on volunteers. By enabling donations, the software includes tiny payments in each OpenAsset transaction:</p>
                                <p>
                                    <ul>
                                        <li>When issuing a new assset, a donation of 0.001 VTC is included</li>
                                        <li>When transferring an asset, a donation of 0.0001 VTC is included</li>
                                    </ul>
                                </p>
                                {donationButtonIntro}
                                <Button className={(this.state.donating === true) ? "btn-danger" : "btn-success"} onClick={(e) => { this.toggleDonations() }}>{(this.state.donating === true) ? "Disable" : "Enable"}</Button>
                            </Col>
                        </Row>
                    </Container>
                );
                break;
        }

        var networkBadge = "";
        if(this.state.network !== "MAINNET") {
            networkBadge = (<Badge color="danger">{this.state.network}</Badge>);
        }
        
        var statusIcon = (<div class="nav-link"> | <MdRefresh /></div>)
        if(this.state.synced) {
            statusIcon = (<div class="nav-link"> | <MdCheckCircle/></div>)
        }

        

        return (
            <div>
                <Navbar color="inverse" light expand="md">
                    <NavbarBrand href="#"  onClick={(e) => { this.setState({page:'home'}) }}><img alt="Logo" src="logo.svg" /> OpenAssets {networkBadge}</NavbarBrand>
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
                                <NavLink href="#" onClick={(e) => { this.setState({page:'donation'}) }}>Donation: <b>{ this.state.donating ? 'On' : 'Off' }</b></NavLink>
                            </NavItem>
                            <NavItem>
                                <div class="nav-link">
                                    | Balance: <b>{ coins.toString() }</b>.<small>{ fractions.toString() }</small>&nbsp;<b>VTC</b>
                                </div>
                            </NavItem>
                            <NavItem>
                                {statusIcon}
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