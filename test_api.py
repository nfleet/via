#!/usr/bin/python
# -*- coding: UTF-8 -*-
import requests, json
from colorama import Fore, Style
import sys, time, simplejson
import argparse
import random

endpoint = '/spp/'
root_url = 'localhost'
PROTOCOL = 'http://'

finland_bbox_boundaries = {
        "long_min": 20.54,         # Border between Swe-Nor-Fin
        "long_max": 31.5867,       # Somewhere in Ilomantsi
        "lat_min": 59.807983,      # Hanko
        "lat_max": 70.092283,      # Nuorgam
}

germany_bbox_boundaries = {
        "long_min": 5.8666667,     # Isenbruch, Nordrhein-Westfalen
        "long_max": 15.033333,     # Deschake, Neißeaue, Saxony
        "lat_min": 47.270108,      # Haldenwanger Eck, Oberstdorf, Bavaria
        "lat_max": 54.9,           # Aventoft, Schleswig-Holstein
}

bboxes = {
        'finland': finland_bbox_boundaries,
        'germany': germany_bbox_boundaries,
}

verbose_treshold = 0

def output(level, s, newline=True):
    if not silent and level <= verbose_treshold:
        if newline:
            sys.stdout.write(s + "\n")
            sys.stdout.flush()
        else:
            sys.stdout.write(s)

def create_matrix(matrix_dim):
    calc = time.time()
    src = [[random.uniform(bbox_boundaries['lat_min'], bbox_boundaries['lat_max']), 
        random.uniform(bbox_boundaries['long_min'], bbox_boundaries['long_max'])] 
        for x in range(matrix_dim)]
    now_calc = time.time() - calc
    output(1, Fore.CYAN + "%dx%d matrix created in %ss. " % (matrix_dim, matrix_dim, str(now_calc)) + Fore.RESET)
    return src

def run_tests(n_tests, matrix_dim, time_out, target_country, speed_profile, print_res, constant_matrix, showmat, matrix=None, url=None,port=80):
    output(1, "Performing " + Fore.MAGENTA + str(n) + Fore.RESET + " tests.")
    output(1, Fore.YELLOW + "Timeout set to %d seconds." % time_out + Fore.RESET)
    output(1, "Selected " + Fore.CYAN + "%s" % country + Fore.RESET + " as country.")
    c = time.time()
    succ, fail, reqfail = 0, 0, 0
    src = None
    for i in range(n_tests):
        # nested ternary... I'm going to hell for this!
        src = matrix if matrix != None else (create_matrix(matrix_dim) if src == None or not constant_matrix else src)
        if showmat:
            output(0, ">>> " + str(src))

        payload = { 'matrix': str(src), 'country': target_country, 'speed_profile': speed_profile }

        # post
        burl = "%s%s:%d%s" % (PROTOCOL, url, port, endpoint)
        r = requests.post(burl, data=json.dumps(payload), timeout=time_out)
        # got created response
        now_req = time.time()
        if r.status_code == 201:
            # wait before polling
            time.sleep(1)
            
            resLoc = r.headers['location']

            new_url = "%s%s:%d%s" % (PROTOCOL, url, port, resLoc)

            # start polling
            done = False
            poll_req = None
            output(1, "Polling... ", False)
            sys.stdout.flush()
            i = 0
            while not done:
                poll_req = requests.get(new_url, allow_redirects=False)
                if poll_req.status_code == 303:
                    done = True
                    break
                elif poll_req.status_code == 502:
                    output(0, "ACHTUNG! Server down... " + Fore.RED + "HALT!")
                    sys.exit(-1)
                try:
                    res = poll_req.json()
                    progress = res['progress']
                    output(1, "%s... " % progress, False)
                    sys.stdout.flush()
                except simplejson.scanner.JSONDecodeError, e:
                    print(e)
                time.sleep(1)

            output(1, " done.")

            assert('location' in poll_req.headers and poll_req.status_code == 303)
            
            result_url = "%s%s:%d%s" % (PROTOCOL, url, port, poll_req.headers['location'])
            result_req = requests.get(result_url)
            if result_req.status_code == 200:
                try:
                    now_req = time.time() - now_req
                    #print result_req.json()
                    status = "%sHTTP 200" % Fore.GREEN
                    status_mat = "%s%dx%d" % (Fore.MAGENTA, matrix_dim, matrix_dim)
                    status_country = "%s%s%s%s" % (Style.BRIGHT, Fore.RED, target_country.upper(), Style.RESET_ALL)
                    status_profile = "%s%s km/h" % (Fore.YELLOW, speed_profile)
                    status_time = "%s%.3f sec%s" % (Fore.CYAN, now_req, Fore.RESET)

                    msg = "%s %s %s %s %s" % (status, status_mat, status_country, status_profile, status_time)
                    output(0, msg)
                    if print_res:
                        output(0, "<<< " + result_req.text)
                except Exception, e:
                    output(0, str(e))
                    output(1, result_req.text)
                    sys.exit(-1)
                succ += 1
        else:
            now_req = time.time() - now_req
            output(0, Fore.MAGENTA + "HTTP %d " % r.status_code + r.text[:30] + " [...] " + Fore.YELLOW +  str(now_req) + "s. TIMEOUT." + Fore.RESET)
            reqfail += 1
    done = time.time() - c

    output(1, Fore.CYAN + "Finished in " + str(done) + "s" + Fore.RESET)
    output(1, Fore.GREEN + ("%d/%d succeeded") % (succ, n) + Fore.RESET + ", " + Fore.RED + ("%d/%d failed.") % (fail, n) + Fore.RESET + ", " + Fore.YELLOW + ("%d timeouts") % reqfail + Fore.RESET)

if __name__=="__main__":
    # create command line arguments
    parser = argparse.ArgumentParser(description="Test the SPP API.", formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument('-u', '--url', action="store", metavar='URL', type=str, help="root url to use", default=root_url)
    parser.add_argument('-r', '--requests', action="store", metavar='N', type=int, help="number of requests to be made to the API", default="1")
    parser.add_argument('-d', '--dim', action="store", metavar="M", type=int, help="create a MxM matrix", default=100)
    parser.add_argument('-t', '--time', action="store", metavar="SEC", type=int, help="request timeout", default=60)
    parser.add_argument('-s', '--silent', action="store_true", help="print no output")
    parser.add_argument('-m', '--show-matrix', action="store_true", help="print the matrix that was calculated")
    parser.add_argument('-c', '--constant', action="store_true", help="keep matrices constant accross multiple requests (only works when profile is ALL)")
    parser.add_argument('-v', '--verbose', action="store_true", help="verbose output")
    parser.add_argument('-p', '--print', action="store_true", help="print API result")
    parser.add_argument('-o', '--port', action="store", metavar='PORT', type=int,help="port to use", default=80)
    parser.add_argument('country', type=str, metavar='COUNTRY', help="select country (finland, germany)")
    parser.add_argument('profile', type=str, metavar='SPEED', help="use profile for km/h speed (40, 60, 80, 100, 120)\nif ALL, then run a test for all profiles")

    if len(sys.argv) <= 1:
        parser.print_help()
        sys.exit(-1)

    bla = parser.parse_args()
    bla = vars(bla)
    country = bla["country"]
    n = bla["requests"]
    prof = bla["profile"]
    dim = bla["dim"]
    t = bla["time"]
    silent = bla["silent"]
    cons = bla["constant"]
    showmat = bla["show_matrix"]
    res = bla["print"]
    url = bla["url"]
    port = bla["port"]

    if bla["verbose"]:
        verbose_treshold = 1

    bbox_boundaries = bboxes[country]

    if prof == "ALL":
        profs = range(40, 140, 20)
        mat = None
        if cons:
            mat = create_matrix(dim)
        for p in profs:
            run_tests(n, dim, t, country, p, res, cons, showmat=showmat, matrix=mat, url=url, port=port)
    else:
        run_tests(n, dim, t, country, int(prof), res, cons, showmat=showmat, url=url, port=port)
